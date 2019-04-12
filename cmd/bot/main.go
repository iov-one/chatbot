package main

import (
	"os"

	"github.com/go-chat-bot/bot/slack"
	"github.com/iov-one/chatbot"
)

const slackTokenEnv = "CHATBOT_SLACK_TOKEN"

func main() {
	commands := []chatbot.Command{
		chatbot.NewDeployCommand(),
		chatbot.NewResetCommand(),
	}

	for _, cmd := range commands {
		cmd.Register()
	}

	if os.Getenv(slackTokenEnv) == "" {
		chatbot.Log("you must supply a slack token via %s env variable\n", slackTokenEnv)
		os.Exit(1)
	}

	slack.Run(os.Getenv(slackTokenEnv))
}
