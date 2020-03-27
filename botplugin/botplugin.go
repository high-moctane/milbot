package botplugin

import (
	"context"

	"github.com/slack-go/slack"
)

// Plugin はプラグインが満たすべきインターフェースです。
// Name はプラグインの名前です。
// Start で起動処理をします。
// Serve で *slack.RTMEvent を受け取って，*slack.Client を用いて返事をするなり
// なんなりします。
// Stop で終了処理をします。
type Plugin interface {
	Start() error
	Serve(context.Context, *slack.Client, slack.RTMEvent) error
	Help() string
	Stop() error
}
