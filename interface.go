package chatbot

import "github.com/go-chat-bot/bot"

// Command represents a command that could be run on slack
type Command interface {
	// Func returns a command that could be run on slack
	Func() func(*bot.Cmd) (string, error)
	// Register registers the func in global slack command registry
	Register()
}
