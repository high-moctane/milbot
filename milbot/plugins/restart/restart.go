package restart

import (
	"log"
	"os"
	"regexp"

	"github.com/high-moctane/milbot/milbot/postlog"

	"github.com/nlopes/slack"
)

// logger はちょっとリッチにしといた
var logger = log.New(os.Stdout, "milbot-restart: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)

// validPrefix は有効な先頭文字列
var validPrefix = regexp.MustCompile(`(?i)^milbot restart`)

// Plugin の中身は必要ない
type Plugin struct{}

// New でプラグインを作成する
func New() Plugin {
	return struct{}{}
}

// Serve では "milbot restart" に反応して終了コード 1 で終了する
func (p Plugin) Serve(api *slack.Client, ch <-chan slack.RTMEvent) {
	for msg := range ch {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			// bot かどうかを判定
			if ev.BotID != "" {
				continue
			}

			if validPrefix.MatchString(ev.Text) {
				restart(api, ev)
			}
		}
	}
}

// Stop は実際なにもしないぞ！
func (p Plugin) Stop() error {
	return nil
}

// restart でログを吐いて終了する
func restart(api *slack.Client, ev *slack.MessageEvent) {
	user, err := api.GetUserInfo(ev.User)
	username := user.Name
	if err != nil {
		username = ""
	}

	postlog.Log("restart: restarted by ", username)
	logger.Printf("restarted by %s", username)
	os.Exit(1)
}
