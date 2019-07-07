package hello

import (
	"log"
	"os"

	"github.com/high-moctane/milbot/milbot/postlog"
	"github.com/nlopes/slack"
)

// logger はちょっとリッチにしといた
var logger = log.New(os.Stdout, "milbot-hello: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)

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
	postlog.Log("hello: received hello")
	logger.Print("received hello")
}
