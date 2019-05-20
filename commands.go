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
	invalidDeploySyntax = "Deploy command requires 4 parameters: " +
		"```!deploy %s your_app your_container your/docker:image``` \nGot: ```!deploy %s```"
	invalidImageFormat = "```Invalid image format, should be your_dockerhub_repo:tag``` \nGot: ```%s```"
	invalidImage       = "```Invalid image, tag %s does not exist in dockerhub repo %s```"
	invalidResetSyntax = "Reset command requires 2 parameters: " +
		"```!deploy %s your_app``` \nGot: ```!deploy %s```"
	appNotFound       = "Sorry, app %s could not be found"
	cmdResponse       = "This is the response to your request:\n ```\n%s\n``` "
	clusterNameNotice = "You must specify cluster name in order to use the command:\n ```!%s %s %s\n```"
)

func wrongClusterName(cmd *bot.Cmd, clusterName string) (string, bool) {
	if len(cmd.Args) > 0 && strings.HasSuffix(cmd.Args[0], "net") {
		return "", cmd.Args[0] != clusterName
	}

	return fmt.Sprintf(clusterNameNotice, cmd.Command, clusterName, strings.Join(cmd.Args, " ")), true
}

type deployCommand struct {
	client      *http.Client
	clusterName string
}

func (c *deployCommand) Func() func(*bot.Cmd) (string, error) {
	panic("stub")
}

func NewDeployCommand(clusterName string) Command {
	return &deployCommand{
		client:      &http.Client{Timeout: 2 * time.Second},
		clusterName: clusterName,
	}
}

func (c *deployCommand) Register() {
	bot.RegisterCommandV3(
		"deploy",
		"Kubectl deployment abstraction",
		fmt.Sprintf("%s your_app your_container your/docker:image", c.clusterName),
		c.Func3())
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
			msg, isWrong := wrongClusterName(cmd, c.clusterName)
			if isWrong {
				res.Message <- msg
				return
			}

			if len(cmd.Args) != 4 {
				res.Message <- fmt.Sprintf(invalidDeploySyntax, c.clusterName, strings.Join(cmd.Args, " "))
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
	lock        sync.Locker
	clusterName string
}

func (c *resetCommand) Func3() func(*bot.Cmd) (bot.CmdResultV3, error) {
	panic("stub")
}

func NewResetCommand(clusterName string) Command {
	return &resetCommand{
		lock:        &sync.Mutex{},
		clusterName: clusterName,
	}
}

func (c *resetCommand) Func() func(*bot.Cmd) (string, error) {
	return func(cmd *bot.Cmd) (s string, e error) {
		c.lock.Lock()
		defer c.lock.Unlock()
		msg, isWrong := wrongClusterName(cmd, c.clusterName)
		if isWrong {
			return msg, nil
		}

		if len(cmd.Args) != 2 {
			return fmt.Sprintf(invalidResetSyntax, c.clusterName, strings.Join(cmd.Args, " ")), nil
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
		fmt.Sprintf("%s your_app", c.clusterName),
		c.Func())
}
