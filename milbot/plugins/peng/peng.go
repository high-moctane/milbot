package peng

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/high-moctane/milbot/milbot/botutils"

	"github.com/nlopes/slack"
)

// peng を発動する先頭文字列
var pengPrefix = regexp.MustCompile(`(?i)^milbot peng`)
var helpPrefix = regexp.MustCompile(`(?i)^milbot peng help`)

// 当たったときに出る追加の絵文字
var jackpotText = ":tada::tada::tada:"

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

			if helpPrefix.MatchString(ev.Text) {
				go help(api, ev)
			} else if pengPrefix.MatchString(ev.Text) {
				go peng(api, ev)
			}
		}
	}
}

// Stop は実際なにもしないぞ！
func (p Plugin) Stop() error {
	return nil
}

func help(api *slack.Client, ev *slack.MessageEvent) {
	botutils.LogEventReceive(api, ev, "peng help")

	jackProb, err := jackpotProbability()
	if err != nil {
		botutils.LogBoth("peng: peng help error: ", err)
		return
	}

	mes := "当たりの確率は\n"
	mes += strconv.FormatFloat(jackProb, 'f', 4, 64) + "\n"
	mes += "です(｀･ω･´):fire::penguin::fire:"

	botutils.SendMessageWithLog(api, ev, mes)
}

// peng はペンギン燃やしを送信する
func peng(api *slack.Client, ev *slack.MessageEvent) {
	botutils.LogEventReceive(api, ev, "peng")

	jackProb, err := jackpotProbability()
	if err != nil {
		botutils.LogBoth("peng: peng error: ", err)
		return
	}

	mes := firePenguin(jackProb)

	botutils.SendMessageWithLog(api, ev, mes)
}

// jackpotProbability は当たりの確率を返す
func jackpotProbability() (float64, error) {
	str, ok := os.LookupEnv("PENG_PROBABILITY")
	if !ok {
		return 0.0, fmt.Errorf("could not find PENG_PROBABILITY")
	}

	prob, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0.0, err
	}

	return prob, nil
}

// 各要素が :fire: になる確率
func fireProbability(jackProb float64) float64 {
	return math.Pow(jackProb, 1.0/8.0)
}

// genEmoji は fireProb の確率で :fire: になる
func genEmoji(fireProb float64) string {
	if rand.Float64() < fireProb {
		return ":fire:"
	}
	return ":snowflake:"
}

// firePenguin は jackProb のもとで確率的ファイアペンギンを生成する
func firePenguin(jackProb float64) string {
	fireProb := fireProbability(jackProb)

	// :penguin: を囲む絵文字を生成する
	fireSnow := make([]string, 8)
	hitCnt := 0
	for i := 0; i < 8; i++ {
		fireSnow[i] = genEmoji(fireProb)
		if fireSnow[i] == ":fire:" {
			hitCnt++
		}
	}

	mes := strings.Join(fireSnow[:3], "") + "\n"
	mes += fireSnow[3] + ":penguin:" + fireSnow[4] + "\n"
	mes += strings.Join(fireSnow[5:], "")
	if hitCnt == 8 {
		mes += "\n" + jackpotText
	}
	return mes
}
