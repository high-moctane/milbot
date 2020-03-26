package plugin

import (
	"context"

	"github.com/slack-go/slack"
)

// Plugin はプラグインが満たすべきインターフェースです。
// Start() で起動処理をします。
// Respond で *slack.RTMEvent を受け取って，*slack.Client を用いて返事をするなり
// なんなりします。
// Stop() で終了処理をします。
type Plugin interface {
	Start() error
	Respond(context.Context, *slack.Client, *slack.RTMEvent) error
	Stop() error
}
