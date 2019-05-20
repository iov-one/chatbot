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
	appNotFound       = "Sorry, app %s could not be found"
	cmdResponse       = "This is the response to your request:\n ```\n%s\n``` "
	clusterNameNotice = "You must specify cluster name in order to use the command:\n ```%s %s %s\n```"
)

func useClusterName(clusterName string) func(cmd *bot.Cmd) (s string, e error) {
	return func(cmd *bot.Cmd) (s string, e error) {
		if len(cmd.Args) > 0 && strings.HasSuffix(cmd.Args[0], "net") {
			return "", nil
		}
		return fmt.Sprintf(clusterNameNotice, cmd.Command, clusterName, strings.Join(cmd.Args, " ")), nil
	}
}

type deployCommand struct {
	client *http.Client
}

func (c *deployCommand) Func() func(*bot.Cmd) (string, error) {
	panic("stub")
}

func NewDeployCommand() Command {
	return &deployCommand{
		client: &http.Client{Timeout: 2 * time.Second},
	}
}

func (c *deployCommand) Register(clusterName string) {
	bot.RegisterCommandV3(
		fmt.Sprintf("deploy %s", clusterName),
		"Kubectl deployment abstraction",
		fmt.Sprintf("%s your_app your_container your/docker:image", clusterName),
		c.Func3())
	bot.RegisterCommand(
		"deploy",
		"Kubectl deployment abstraction",
		"",
		useClusterName(clusterName))
}

func (c *deployCommand) Func3() func(*bot.Cmd) (bot.CmdResultV3, error) {
	commandString := "kubectl set image %s %s %s=%s"
	statusCommandString := "kubectl get pods --selector=app=%s"
	return func(cmd *bot.Cmd) (s bot.CmdResultV3, e error) {
		res := bot.CmdResultV3{
			Message: make(chan string, 1),
			Done:    make(chan bool, 1),
		}

		go func() {
			defer func() {
				res.Done <- true
			}()
			if len(cmd.Args) != 3 {
				res.Message <- fmt.Sprintf(invalidDeploySyntax, strings.Join(cmd.Args, " "))
				return
			}

			app := cmd.Args[0]
			container := cmd.Args[1]
			image := cmd.Args[2]

			imageParts := strings.Split(image, ":")
			if len(imageParts) != 2 {
				res.Message <- fmt.Sprintf(invalidImageFormat, image)
				return
			}

			imageRepo := imageParts[0]
			imageTag := imageParts[1]

			if r, err :=
				c.client.Get(fmt.Sprintf(
					"https://index.docker.io/v1/repositories/%s/tags/%s",
					imageRepo, imageTag)); r.StatusCode != 200 || err != nil {
				res.Message <- fmt.Sprintf(invalidImage, imageTag, imageRepo)
				return
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
				res.Message <- fmt.Sprintf(appNotFound, app)
				return
			}

			res.Message <- fmt.Sprintf(cmdResponse, output)
			time.Sleep(time.Second * 20)
			res.Message <- fmt.Sprintf(cmdResponse, execute(fmt.Sprintf(statusCommandString, app)))

			return
		}()

		return res, nil
	}
}

type resetCommand struct {
	lock sync.Locker
}

func (c *resetCommand) Func3() func(*bot.Cmd) (bot.CmdResultV3, error) {
	panic("stub")
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

func (c *resetCommand) Register(clusterName string) {
	bot.RegisterCommand(
		fmt.Sprintf("reset %s", clusterName),
		"Kubectl reset abstraction to allow removing pvc for stateful sets by app label and recreating them",
		fmt.Sprintf("%s your_app", clusterName),
		c.Func())
	bot.RegisterCommand(
		"reset",
		"Kubectl reset abstraction",
		"",
		useClusterName(clusterName))
}
