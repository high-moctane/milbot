package exit

import (
	"log"
	"os"
	"regexp"

	"github.com/nlopes/slack"
)

// logger はちょっとリッチにしといた
var logger = log.New(os.Stdout, "milbot-exit: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)

// validPrefix は有効な先頭文字列
var validPrefix = regexp.MustCompile(`(?i)^milbot exit`)

// Plugin の中身は必要ない
type Plugin struct{}

// New でプラグインを作成する
func New() Plugin {
	return struct{}{}
}

// Serve では "milbot exit" に反応して終了コード 0 で終了する
func (p Plugin) Serve(api *slack.Client, ch chan slack.RTMEvent) {
	for msg := range ch {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			// bot かどうかを判定
			if ev.BotID != "" {
				continue
			}

			if validPrefix.MatchString(ev.Text) {
				exit(api, ev)
			}
		}
	}
}

// Stop は実際なにもしないぞ！
func (p Plugin) Stop() error {
	return nil
}

// exit でログを吐いて終了する
func exit(api *slack.Client, ev *slack.MessageEvent) {
	user, err := api.GetUserInfo(ev.User)
	username := user.Name
	if err != nil {
		username = ""
	}

	logger.Printf("exited by %s", username)
	os.Exit(0)
}
