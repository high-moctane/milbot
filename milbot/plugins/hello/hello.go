package hello

import (
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
func (p Plugin) Serve(api *slack.Client, ch <-chan slack.RTMEvent) {
	for msg := range ch {
		switch msg.Data.(type) {
		case *slack.HelloEvent:
			go hello()
		}
	}
}

// Stop は実際なにもしないぞ！
func (p Plugin) Stop() error {
	return nil
}

// hello は接続したよのログを出す
func hello() {
	botutils.LogBoth("hello: received hello")
}
