package ping

import (
	"regexp"

	"github.com/high-moctane/milbot/milbot/botutils"
	"github.com/nlopes/slack"
)

// validPrefix は有効な先頭文字列
var validPrefix = regexp.MustCompile(`(?i)^milbot ping`)

// Plugin の中身は必要ない
type Plugin struct{}

// New でプラグインを作成する
func New() Plugin {
	return struct{}{}
}

// Serve では "milbot ping" に反応して "pong(｀･ω･´)" と返します
func (p Plugin) Serve(api *slack.Client, ch <-chan slack.RTMEvent) {
	for msg := range ch {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			// bot かどうかを判定
			if ev.BotID != "" {
				continue
			}

			if validPrefix.MatchString(ev.Text) {
				botutils.LogEventReceive(api, ev, "ping")
				botutils.SendMessageWithLog(api, ev.Channel, "pong(｀･ω･´)")
			}
		}
	}
}

// Stop は実際なにもしないぞ！
func (p Plugin) Stop() error {
	return nil
}
