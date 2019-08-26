package exit

import (
	"context"
	"os"
	"regexp"

	"github.com/high-moctane/milbot/milbot/botutils"
	"github.com/nlopes/slack"
)

// validPrefix は有効な先頭文字列
var validPrefix = regexp.MustCompile(`(?i)^milbot exit`)

// Plugin の中身は必要ない
type Plugin struct{}

// New でプラグインを作成する
func New() Plugin {
	return struct{}{}
}

// Serve では "milbot exit" に反応して終了コード 0 で終了する
func (p Plugin) Serve(ctx context.Context, api *slack.Client, ch <-chan slack.RTMEvent) {
	for {
		select {
		case <-ctx.Done():
			return

		case msg := <-ch:
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
}

// exit でログを吐いて終了する
func exit(api *slack.Client, ev *slack.MessageEvent) {
	user, err := api.GetUserInfo(ev.User)
	username := user.Name
	if err != nil {
		username = ""
	}

	botutils.Log("exit: exited by ", username)
	os.Exit(0)
}
