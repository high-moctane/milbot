package plugin

import (
	"context"

	"github.com/slack-go/slack"
)

// Plugin はプラグインが満たすべきインターフェースです。Respond で *slack.RTMEvent を
// 受け取って，*slack.Client を用いて返事をするなりなんなりします。
// Stop() で終了処理をします。
type Plugin interface {
	Respond(context.Context, *slack.Client, *slack.RTMEvent)
	Stop()
}
