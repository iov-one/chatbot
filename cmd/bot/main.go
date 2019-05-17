package main

import (
	"os"
	"strings"

	"github.com/go-chat-bot/bot/slack"
	"github.com/iov-one/chatbot"
)

const slackTokenEnv = "CHATBOT_SLACK_TOKEN"
const clusterName = "CHATBOT_CLUSTER_NAME"

func main() {
	commands := []chatbot.Command{
		chatbot.NewDeployCommand(),
		chatbot.NewResetCommand(),
	}

	if !strings.HasSuffix(os.Getenv(clusterName), "net") {
		chatbot.Log("you must supply a clusterName via %s env variable and it has to end in 'net', like 'devnet'\n",
			clusterName)
		os.Exit(1)
	}

	if os.Getenv(slackTokenEnv) == "" {
		chatbot.Log("you must supply a slack token via %s env variable\n", slackTokenEnv)
		os.Exit(1)
	}

	for _, cmd := range commands {
		cmd.Register(os.Getenv(clusterName))
	}

	slack.Run(os.Getenv(slackTokenEnv))
}
