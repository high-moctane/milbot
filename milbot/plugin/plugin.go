package plugin

import "github.com/nlopes/slack"

// Plugin を実装すると Plugin になれるよ(｀･ω･´)
type Plugin interface {
	Serve(*slack.Client, <-chan slack.RTMEvent)
	Stop() error
}
