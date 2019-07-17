package kitakunoki

import (
	"math/rand"
	"regexp"
	"strconv"
	"time"

	"github.com/high-moctane/milbot/milbot/botutils"
	"github.com/nlopes/slack"
)

func init() {
	// rand を使うので seed を設定する
	rand.Seed(time.Now().UnixNano())
}

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
		botutils.LogBoth("kitakunoki: could not fetch html: ", err)
		p := &Plugin{ok: false}
		return p
	}

	entries, err := parseHTML(r)
	if err != nil {
		botutils.LogBoth("kitakunoki: could not parse html: ", err)
		p := &Plugin{ok: false}
		return p
	}

	p := &Plugin{entries: entries, ok: true}
	return p
}

// Serve では kitakunoki を発見する
func (p *Plugin) Serve(api *slack.Client, ch <-chan slack.RTMEvent) {
	// 帰宅のアラートサーバーを立てる
	go kitakunoAlertServer(api, p)

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

// help のメッセージを送信する
func help(api *slack.Client, ev *slack.MessageEvent) {
	botutils.LogEventReceive(api, ev, "kitakunoki help")

	mes := "`帰宅の木` に反応して今日の帰宅の木をお知らせします。\n" +
		"また `帰宅の木` が含まれるメッセージに絵文字をつけます(｀･ω･´)"

	botutils.SendMessageWithLog(api, ev.Channel, mes)
}

// kitakunonae を植える
func kitakunonae(api *slack.Client, ev *slack.MessageEvent) {
	botutils.LogEventReceive(api, ev, "kitakunonae")

	err := api.AddReaction("seedling", slack.ItemRef{
		Channel:   ev.Channel,
		Timestamp: ev.Timestamp,
	})
	if err != nil {
		botutils.LogBoth("kitakunoki: kitakunokinonae: fail add reaction")
		return
	}
	botutils.Logf("kitakunokinonae: added reaction to {channel: %s, ts: %s}", ev.Channel, ev.Timestamp)
}

// kitakunoki を生やす
func (p *Plugin) kitakunoki(api *slack.Client, ev *slack.MessageEvent) {
	botutils.LogEventReceive(api, ev, "kitakunoki")

	if err := kitakunoAddReaction(api, ev); err != nil {
		botutils.Log("kitakunoki: kitakunoki: failed add reaction: ", err)
		// ここで return はしない
	}

	botutils.SendMessageWithLog(api, ev.Channel, p.kitakunoMessage())
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
	botutils.Logf("add reaction to {channel: %s, ts: %s}", ev.Channel, ev.Timestamp)
	return nil
}

// kitakunoMessage は今日の帰宅の木のメッセージを構築する
func (p *Plugin) kitakunoMessage() string {
	ki := p.todaysChoice()
	return "今日の帰宅の木は " + ki.name + "です(｀･ω･´)\n" + ki.url
}
