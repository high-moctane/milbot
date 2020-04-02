package botplugin

import (
	"context"

	"github.com/slack-go/slack"
)

// Plugin はプラグインが満たすべきインターフェースです。
// Name はプラグインの名前です。
// Start で起動処理をします。必要であれば *slack.Client を保存してください。
// Serve で *slack.RTMEvent を受け取って返事をするなりします。
// Stop で終了処理をします。
// Help で使い方を説明したメッセージを返します。
type Plugin interface {
	Start(*slack.Client) error
	Serve(context.Context, slack.RTMEvent) error
	Stop() error
	Help() string
}
