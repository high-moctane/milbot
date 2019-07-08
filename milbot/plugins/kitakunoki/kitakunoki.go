package kitakunoki

import (
	"log"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/high-moctane/milbot/milbot/postlog"
	"github.com/nlopes/slack"
)

func init() {
	// rand を使うので seed を設定する
	rand.Seed(time.Now().UnixNano())
}

// logger はちょっとリッチにしといた
var logger = log.New(os.Stdout, "milbot-kitakunoki: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)

// atnd を発動する先頭文字列
var kitakunokiRegexp = regexp.MustCompile(`(帰宅|きたく)の(木|き)`)
var kitakunonaeRegexp = regexp.MustCompile(`(帰宅|きたく)の(木|き)の(苗|なえ)`)
var helpPrefix = regexp.MustCompile(`(?i)^milbot kitakunoki help`)

// 木の絵文字
var treeEmojis = []string{"palm_tree", "evergreen_tree", "deciduous_tree", "christmas_tree"}

// Entry は帰宅の木の名前と url です
type entry struct {
	name, url string
}

// Plugin の中に帰宅の木のリストを作っておく
type Plugin struct {
	entries []entry // 帰宅の木のリスト
	ok      bool    // ちゃんとリストが読み込めたかどうか
}

// New でプラグインを作成する
func New() *Plugin {
	r, err := fetchHTML()
	if err != nil {
		postlog.Log("kitakunoki: could not fetch html: ", err)
		logger.Print("kitakunoki: could not fetch html: ", err)
		p := &Plugin{ok: false}
		return p
	}

	entries, err := parseHTML(r)
	if err != nil {
		postlog.Log("kitakunoki: could not parse html: ", err)
		logger.Print("kitakunoki: could not parse html: ", err)
		p := &Plugin{ok: false}
		return p
	}

	p := &Plugin{entries: entries, ok: true}
	return p
}

// Serve では kitakunoki を発見する
func (p *Plugin) Serve(api *slack.Client, ch <-chan slack.RTMEvent) {
	for msg := range ch {
		// コンストラクトに失敗したらできることがないので終了
		if !p.ok {
			return
		}

		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			// bot かどうかを判定
			if ev.BotID != "" {
				continue
			}

			if helpPrefix.MatchString(ev.Text) {
				go help(api, ev)
			} else if kitakunonaeRegexp.MatchString(ev.Text) {
				go kitakunonae(api, ev)
			} else if kitakunokiRegexp.MatchString(ev.Text) {
				go p.kitakunoki(api, ev)
			}
		}
	}
}

// Stop は実際なにもしないぞ！
func (p *Plugin) Stop() error {
	return nil
}

// choice は帰宅の木から今日の木をランダムに選ぶ
func (p *Plugin) todaysChoice() entry {
	if len(p.entries) < 0 {
		panic("kitakunoki empty entries")
	}

	seedstr := time.Now().Format("20060102")
	seed, _ := strconv.Atoi(seedstr)
	rnd := rand.New(rand.NewSource(int64(seed)))

	idx := rnd.Intn(len(p.entries))
	return p.entries[idx]
}

// receiveLog でメッセージを受けっとたよーというログを吐く
func receiveLog(api *slack.Client, ev *slack.MessageEvent, mes string) {
	user, err := api.GetUserInfo(ev.User)
	username := user.Name
	if err != nil {
		username = ""
	}
	logger.Print("received "+mes+" by ", username)
}

// sendLog でメッセージを送ったよーというログを吐く
func sendLog(channel, ts, text string) {
	logger.Printf("sent message: {channel: %s, ts: %s, text: %s}", channel, ts, text)
}

// postMessage でメッセージを送信する（ログ付き！）
func postMessage(api *slack.Client, ev *slack.MessageEvent, mes string) {
	channel, ts, text, err := api.SendMessage(
		ev.Channel,
		slack.MsgOptionText(mes, true),
	)
	if err != nil {
		postlog.Log("kitakunoki: ", err)
		logger.Print(err)
		return
	}
	sendLog(channel, ts, text)
}

// help のメッセージを送信する
func help(api *slack.Client, ev *slack.MessageEvent) {
	receiveLog(api, ev, "kitakunoki help")

	mes := "`帰宅の木` に反応して今日の帰宅の木をお知らせします。\n" +
		"また `帰宅の木` が含まれるメッセージに絵文字をつけます(｀･ω･´)"

	postMessage(api, ev, mes)
}

// kitakunonae を植える
func kitakunonae(api *slack.Client, ev *slack.MessageEvent) {
	receiveLog(api, ev, "kitakunokinonae")

	err := api.AddReaction("seedling", slack.ItemRef{
		Channel:   ev.Channel,
		Timestamp: ev.Timestamp,
	})
	if err != nil {
		postlog.Log("kitakunoki: kitakunokinonae: fail add reaction")
		logger.Printf("kitakunoki: kitakunokinonae: fail add reaction to {channel: %s, ts: %s}", ev.Channel, ev.Timestamp)
		return
	}
	logger.Printf("kitakunokinonae: added reaction to {channel: %s, ts: %s}", ev.Channel, ev.Timestamp)
}

// kitakunoki を生やす
func (p *Plugin) kitakunoki(api *slack.Client, ev *slack.MessageEvent) {
	receiveLog(api, ev, "kitakunoki")

	if err := kitakunoAddReaction(api, ev); err != nil {
		postlog.Log("kitakunoki: kitakunoki: failed add reaction: ", err)
		logger.Print("failed add reaction: ", err)
		// ここで return はしない
	}

	postMessage(api, ev, p.kitakunoMessage())
}

// kitakunoAddReaction は適当な木を生やす
func kitakunoAddReaction(api *slack.Client, ev *slack.MessageEvent) error {
	idx := rand.Intn(len(treeEmojis))
	emoji := treeEmojis[idx]

	err := api.AddReaction(emoji, slack.ItemRef{
		Channel:   ev.Channel,
		Timestamp: ev.Timestamp,
	})
	if err != nil {
		return err
	}
	logger.Printf("add reaction to {channel: %s, ts: %s}", ev.Channel, ev.Timestamp)
	return nil
}

// kitakunoMessage は今日の帰宅の木のメッセージを構築する
func (p *Plugin) kitakunoMessage() string {
	ki := p.todaysChoice()
	return "今日の帰宅の木は " + ki.name + "です(｀･ω･´)\n" + ki.url
}
