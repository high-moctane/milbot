package script

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/high-moctane/milbot/milbot/botutils"

	"github.com/nlopes/slack"
)

// atnd を発動する先頭文字列
var helpPrefix = regexp.MustCompile(`(?i)^milbot script help`)
var bashPrefix = regexp.MustCompile(`(?i)^milbot bash`)
var python3Prefix = regexp.MustCompile(`(?i)^milbot python3`)

// Plugin の中身は必要ない
type Plugin struct{}

// New でプラグインを作成する
func New() Plugin {
	return struct{}{}
}

// Serve では atnd のクエリを振り分ける
func (p Plugin) Serve(api *slack.Client, ch <-chan slack.RTMEvent) {
	for msg := range ch {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			// bot かどうかを判定
			// 謎のコメントアウト
			// if ev.BotID != "" {
			// 	continue
			// }

			if helpPrefix.MatchString(ev.Text) {
				go help(api, ev)
			} else if bashPrefix.MatchString(ev.Text) {
				go bash(api, ev)
			} else if python3Prefix.MatchString(ev.Text) {
				go python3(api, ev)
			}
		}
	}
}

// Stop は実際なにもしないぞ！
func (p Plugin) Stop() error {
	return nil
}

// help のメッセージを送信する
func help(api *slack.Client, ev *slack.MessageEvent) {
	botutils.LogEventReceive(api, ev, "atnd help")

	mes := "" // TODO

	botutils.SendMessageWithLog(api, ev, mes)
}

// bash を走らせる
func bash(api *slack.Client, ev *slack.MessageEvent) {
	input, err := parseInput(ev.Text)
	if err != nil {
		botutils.SendMessageWithLog(api, ev, "入力フォーマットが不正です(´･ω･｀)")
	}
	botutils.SendMessageWithLog(api, ev, input)
	// out, err := run([]string{"bash", "main.sh"}, "main.sh", input)
	// if err != nil {
	// 	botutils.SendMessageWithLog(api, ev, "実行に失敗しました(´; ω ;｀)")
	// 	botutils.LogBoth("script: bash error: ", err)
	// 	return
	// }
	// botutils.SendMessageWithLog(api, ev, out)
}

// python3 を走らせる
func python3(api *slack.Client, ev *slack.MessageEvent) (string, error) {
	// input := parseInput(ev.Text)
	// out, err := run([]string{"python3", "main.py"}, "main.py", input)
	// if err != nil {
	// 	botutils.SendMessageWithLog(api, ev, "実行に失敗しました(´; ω ;｀)")
	// 	botutils.LogBoth("script: bash error: ", err)
	// 	return
	// }
	// botutils.SendMessageWithLog(api, ev, out)
	return "", nil
}

// parseInput は送られたメッセージからスクリプトの部分を取り出す
func parseInput(text string) (string, error) {
	// TODO: regexp を関数の外で定義する
	r := regexp.MustCompile(`\s.`)
	indices := r.FindAllIndex([]byte(text), 2)
	if len(indices) < 2 {
		return "", fmt.Errorf("not found body")
	}

	body := text[indices[1][1]:]
	body = strings.TrimPrefix(body, "```")
	body = strings.TrimSuffix(body, "```")
	return body, nil
}
