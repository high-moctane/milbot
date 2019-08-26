package help

import (
	"context"
	"regexp"

	"github.com/high-moctane/milbot/milbot/botutils"
	"github.com/nlopes/slack"
)

// help を発動する先頭文字列
var helpPrefix = regexp.MustCompile(`(?i)^milbot help`)

var helpMessage = "以下のコマンドを受け付けています。\n" +
	"    `milbot atnd`\n" +
	"    `milbot atnd add`\n" +
	"    `milbot atnd delete`\n" +
	"    `milbot atnd help`\n" +
	"    `milbot atnd list`\n" +
	"    `milbot exit`\n" +
	"    `milbot exit help`\n" +
	"    `milbot help`\n" +
	"    `milbot kitakunoki help`\n" +
	"    `milbot peng`\n" +
	"    `milbot peng help`\n" +
	"    `milbot ping`\n" +
	"    `milbot ping help`\n" +
	"    `milbot restart`\n" +
	"    `milbot restart help`\n" +
	"    `milbot script help`\n" +
	"    `milbot bash`\n" +
	"    `milbot python3`\n" +
	"    `milbot verse help`\n" +
	"\n" +
	"また以下の機能があります。\n" +
	"    帰宅の木\n" +
	"    575 警察\n"

// Plugin の中身は必要ない
type Plugin struct{}

// New でプラグインを作成する
func New() Plugin {
	return struct{}{}
}

// Serve では "milbot help" に反応して help を返す
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

				if helpPrefix.MatchString(ev.Text) {
					go help(api, ev)
				}
			}
		}
	}
}

func help(api *slack.Client, ev *slack.MessageEvent) {
	botutils.SendMessageWithLog(api, ev.Channel, helpMessage)
}
