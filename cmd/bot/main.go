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

	envCluster := os.Getenv(clusterName)

	if !strings.HasSuffix(envCluster, "net") || len(envCluster) < 4 {
		chatbot.Log("you must supply a clusterName via %s env variable and it has to end in 'net' and be "+
			"at least 4 characters long, like 'devnet'\n",
			clusterName)
		os.Exit(1)
	}

	if os.Getenv(slackTokenEnv) == "" {
		chatbot.Log("you must supply a slack token via %s env variable\n", slackTokenEnv)
		os.Exit(1)
	}

	commands := []chatbot.Command{
		chatbot.NewDeployCommand(envCluster),
		chatbot.NewResetCommand(envCluster),
		chatbot.NewP2PCommand(envCluster),
	}

	for _, cmd := range commands {
		cmd.Register()
	}

	slack.Run(os.Getenv(slackTokenEnv))
}
