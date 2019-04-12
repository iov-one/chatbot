package chatbot

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/go-chat-bot/bot"
)

const (
	invalidDeploySyntax = "Deploy command requires 3 parameters: " +
		"```!deploy your_app your_container your/docker:image``` \nGot: ```!deploy %s```"
	invalidResetSyntax = "Reset command requires 1 parameter: " +
		"```!deploy your_app``` \nGot: ```!deploy %s```"
	appNotFound = "Sorry, app %s could not be found"
	cmdResponse = "This is the response to your request:\n ```\n%s\n``` "
)

type deployCommand struct {
}

func NewDeployCommand() Command {
	return &deployCommand{}
}

func (c *deployCommand) Register() {
	bot.RegisterCommand(
		"deploy",
		"Kubectl deployment abstraction",
		"your_app your_container your/docker:image",
		c.Func())
}

func (c *deployCommand) Func() func(*bot.Cmd) (string, error) {
	commandString := "kubectl set image %s %s %s=%s"
	return func(cmd *bot.Cmd) (s string, e error) {
		if len(cmd.Args) != 3 {
			return fmt.Sprintf(invalidDeploySyntax, strings.Join(cmd.Args, " ")), nil
		}

		app := cmd.Args[0]
		container := cmd.Args[1]
		image := cmd.Args[2]

		output := ""
		// Note that this is a hack to make sure we force redeployment even if the image tag is the same
		for _, imageName := range []string{"dummy", image} {
			for _, entityType := range []string{"deployment", "statefulset"} {
				output = execute(fmt.Sprintf(commandString, entityType, app, container, imageName))

				if !isNotFound(output) {
					break
				}
			}
		}

		if isNotFound(output) {
			return fmt.Sprintf(appNotFound, app), nil
		}

		return fmt.Sprintf(cmdResponse, output), nil
	}
}

type resetCommand struct {
	lock sync.Locker
}

func NewResetCommand() Command {
	return &resetCommand{
		lock: &sync.Mutex{},
	}
}

func (c *resetCommand) Func() func(*bot.Cmd) (string, error) {
	return func(cmd *bot.Cmd) (s string, e error) {
		c.lock.Lock()
		defer c.lock.Unlock()
		if len(cmd.Args) != 1 {
			return fmt.Sprintf(invalidResetSyntax, strings.Join(cmd.Args, " ")), nil
		}

		app := cmd.Args[0]
		output, err := c.executeSequence(app)
		if err != nil {
			return "", err
		}

		if isNotFound(output) {
			return fmt.Sprintf(appNotFound, app), nil
		}

		return fmt.Sprintf(cmdResponse, output), nil
	}
}

func (c *resetCommand) executeSequence(app string) (string, error) {
	filename := "/tmp/st.yaml"
	pvcDeleteCmd := "kubectl delete pvc -l app=%s"
	statefulSetStoreCmd := "kubectl get statefulsets.apps %s -o=yaml"
	statefulSetDeleteCmd := "kubectl delete -f /tmp/st.yaml"
	statefulSetApplyCmd := "kubectl apply -f /tmp/st.yaml"

	defer func() {
		_ = os.Remove(filename)
	}()

	output := make([]string, 0)

	statefulSetYaml := execute(fmt.Sprintf(statefulSetStoreCmd, app))
	if isNotFound(statefulSetYaml) {
		return statefulSetYaml, nil
	}

	f, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	_, err = f.WriteString(statefulSetYaml)
	if err != nil {
		return "", err
	}

	output = append(output, execute(statefulSetDeleteCmd))
	output = append(output, execute(fmt.Sprintf(pvcDeleteCmd, app)))
	output = append(output, execute(statefulSetApplyCmd))

	return strings.Join(output, "\n"), nil
}

func (c *resetCommand) Register() {
	bot.RegisterCommand(
		"reset",
		"Kubectl reset abstraction to allow removing pvc for stateful sets by app label and recreating them",
		"your_app",
		c.Func())
}
