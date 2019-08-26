package restart

import (
	"context"
	"os"
	"regexp"

	"github.com/high-moctane/milbot/milbot/botutils"
	"github.com/nlopes/slack"
)

// validPrefix は有効な先頭文字列
var validPrefix = regexp.MustCompile(`(?i)^milbot restart`)

// Plugin の中身は必要ない
type Plugin struct{}

// New でプラグインを作成する
func New() Plugin {
	return struct{}{}
}

// Serve では "milbot restart" に反応して終了コード 1 で終了する
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
					restart(api, ev)
				}
			}
		}
	}
}

// restart でログを吐いて終了する
func restart(api *slack.Client, ev *slack.MessageEvent) {
	user, err := api.GetUserInfo(ev.User)
	username := user.Name
	if err != nil {
		username = ""
	}

	botutils.LogBoth("restart: restarted by ", username)
	os.Exit(1)
}
