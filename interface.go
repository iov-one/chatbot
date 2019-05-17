package chatbot

import "github.com/go-chat-bot/bot"

// Command represents a command that could be run on slack
type Command interface {
	// Func returns a command that could be run on slack
	Func() func(*bot.Cmd) (string, error)
	// Func3 returns an async command that should run on slack
	Func3() func(*bot.Cmd) (bot.CmdResultV3, error)
	// Register registers the func in global slack command registry
	Register(clusterName string)
}
