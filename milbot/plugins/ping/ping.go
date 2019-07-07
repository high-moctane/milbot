package ping

import (
	"log"
	"os"
	"regexp"

	"github.com/nlopes/slack"
)

// logger はちょっとリッチにしといた
var logger = log.New(os.Stdout, "milbot-ping: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)

// validPrefix は有効な先頭文字列
var validPrefix = regexp.MustCompile(`(?i)^milbot ping`)

// Plugin の中身は必要ない
type Plugin struct{}

// New でプラグインを作成する
func New() Plugin {
	return struct{}{}
}

// Serve では "milbot ping" に反応して "pong(｀･ω･´)" と返します
func (p Plugin) Serve(api *slack.Client, ch chan slack.RTMEvent) {
	for msg := range ch {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			// bot かどうかを判定
			if ev.BotID != "" {
				continue
			}

			if validPrefix.MatchString(ev.Text) {
				receiveLog(api, ev)
				sendPong(api, ev.Channel)
			}
		}
	}
}

// Stop は実際なにもしないぞ！
func (p Plugin) Stop() error {
	return nil
}

// sendPong で pong を送る
func sendPong(api *slack.Client, channel string) {
	ch, ts, text, err := api.SendMessage(
		channel,
		slack.MsgOptionText("pong(｀･ω･´)", false),
	)
	if err != nil {
		logger.Printf("send pong error: %s", err)
	}
	// この text が "" なのどうにも納得がいかない
	logger.Printf("send pong: {chan: %s, ts: %s, text: %s}", ch, ts, text)
}

// receiveLog でメッセージを受けっとたよーというログを吐く
func receiveLog(api *slack.Client, ev *slack.MessageEvent) {
	user, err := api.GetUserInfo(ev.User)
	username := user.Name
	if err != nil {
		username = ""
	}
	logger.Print("received ping by ", username)
}
