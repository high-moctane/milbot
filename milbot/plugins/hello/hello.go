package hello

import (
	"context"

	"github.com/high-moctane/milbot/milbot/botutils"
	"github.com/nlopes/slack"
)

// Plugin の中身は必要ない
type Plugin struct{}

// New でプラグインを作成する
func New() Plugin {
	return struct{}{}
}

// Serve では RTM 接続時に ログを出す
func (p Plugin) Serve(ctx context.Context, api *slack.Client, ch <-chan slack.RTMEvent) {
	for {
		select {
		case <-ctx.Done():
			return

		case msg := <-ch:
			switch msg.Data.(type) {
			case *slack.HelloEvent:
				go hello()
			}
		}
	}
}

// hello は接続したよのログを出す
func hello() {
	botutils.LogBoth("hello: received hello")
}
