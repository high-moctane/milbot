package verse

import (
	"regexp"
	"strings"

	"github.com/high-moctane/milbot/milbot/botutils"
	"github.com/nlopes/slack"
)

// help を発動する先頭文字列
var helpPrefix = regexp.MustCompile(`(?i)^milbot verse help`)

// jiritsugo で始まる句かどうかで文頭になれるかどうかを判定する
var jiritsugo = []string{"動詞", "形容詞", "形容動詞", "名詞", "連体詞", "副詞", "接続詞", "感動詞", "フィラー"}

// omitPartOfSpeech の品詞は 575 において無効であると判断する
var omitPartOfSpeech = []string{"記号"}

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
			if ev.BotID != "" {
				continue
			}

			if helpPrefix.MatchString(ev.Text) {
				go help(api, ev)
			}
			go run(api, ev)
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

	mes := "575，57577，7775 に反応します(｀･ω･´)"

	botutils.SendMessageWithLog(api, ev, mes)
}

// run は text から 575 などを探し出して指摘します
func run(api *slack.Client, ev *slack.MessageEvent) {
	senryu := findSenryu(ev.Text)
	tanka := findTanka(ev.Text)
	dodoitsu := findDodoitsu(ev.Text)

	mes := []string{}
	if len(senryu) > 0 {
		mes = append(mes, "Found 575:cop:\n"+strings.Join(senryu, "\n"))
	}
	if len(tanka) > 0 {
		mes = append(mes, "Found 57577:cop:\n"+strings.Join(tanka, "\n"))
	}
	if len(dodoitsu) > 0 {
		mes = append(mes, "Found 7775:cop:\n"+strings.Join(dodoitsu, "\n"))
	}
	if len(mes) == 0 {
		return
	}

	botutils.SendMessageWithLog(api, ev, strings.Join(mes, "\n\n"))
}

// findSenryu は text から 575 を探して返します
func findSenryu(text string) []string {
	return find(text, []int{5, 7, 5}, []bool{true, false, false})
}

// findTanka は text から 57577 を探して返します
func findTanka(text string) []string {
	return find(text, []int{5, 7, 5, 7, 7}, []bool{true, false, false, true, false})
}

// findDodoitsu は text から 7775 を探して返します
func findDodoitsu(text string) []string {
	return find(text, []int{7, 7, 7, 5}, []bool{true, false, false, false})
}
