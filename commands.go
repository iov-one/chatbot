package chatbot

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-chat-bot/bot"
)

const (
	invalidDeploySyntax = "Deploy command requires 3 parameters: " +
		"```!deploy your_app your_container your/docker:image``` \nGot: ```!deploy %s```"
	invalidImageFormat = "```Invalid image format, should be your_dockerhub_repo:tag``` \nGot: ```%s```"
	invalidImage       = "```Invalid image, tag %s does not exist in dockerhub repo %s```"
	invalidResetSyntax = "Reset command requires 1 parameter: " +
		"```!deploy your_app``` \nGot: ```!deploy %s```"
	appNotFound = "Sorry, app %s could not be found"
	cmdResponse = "This is the response to your request:\n ```\n%s\n``` "
)

type deployCommand struct {
	client *http.Client
}

func NewDeployCommand() Command {
	return &deployCommand{
		client: &http.Client{Timeout: 2 * time.Second},
	}
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

		imageParts := strings.Split(image, ":")
		if len(imageParts) != 2 {
			return fmt.Sprintf(invalidImageFormat, image), nil
		}

		imageRepo := imageParts[0]
		imageTag := imageParts[1]

		if r, err :=
			c.client.Get(fmt.Sprintf(
				"https://index.docker.io/v1/repositories/%s/tags/%s",
				imageRepo, imageTag)); r.StatusCode != 200 || err != nil {
			return fmt.Sprintf(invalidImage, imageTag, imageRepo), nil
		}

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
