package plugin

import (
	"context"

	"github.com/nlopes/slack"
)

// Server を実装すると Plugin になれるよ(｀･ω･´)
type Server interface {
	Serve(context.Context, *slack.Client, <-chan slack.RTMEvent)
}
