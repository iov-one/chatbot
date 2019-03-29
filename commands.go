package chatbot

import (
	"fmt"
	"strings"

	"github.com/go-chat-bot/bot"
)

const (
	invalidDeploySyntax = "Deploy command requires 3 parameters: " +
		"```!deploy your_app your_container your/docker:image``` \nGot: ```!deploy %s```"
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
	commandString := "kubectl set image %s %s=%s"
	return func(cmd *bot.Cmd) (s string, e error) {
		if len(cmd.Args) != 3 {
			return fmt.Sprintf(invalidDeploySyntax, strings.Join(cmd.Args, " ")), nil
		}

		app := cmd.Args[0]
		container := cmd.Args[1]
		image := cmd.Args[2]

		output := ""

		for _, entityType := range []string{"deployment", "statefulset"} {
			output := execute(fmt.Sprintf(commandString, entityType, app, container, image))

			if !isNotFound(output) {
				break
			}
		}

		if isNotFound(output) {
			return fmt.Sprintf(appNotFound, app), nil
		}

		return fmt.Sprintf(cmdResponse, output), nil
	}
}
