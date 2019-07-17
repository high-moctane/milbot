package atnd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/high-moctane/milbot/milbot/botutils"

	"github.com/go-redis/redis"
	"github.com/nlopes/slack"
)

// atnd を発動する先頭文字列
var atndPrefix = regexp.MustCompile(`(?i)^milbot atnd`)
var addPrefix = regexp.MustCompile(`(?i)^milbot atnd add`)
var removePrefix = regexp.MustCompile(`(?i)^milbot atnd delete`)
var listPrefix = regexp.MustCompile(`(?i)^milbot atnd list`)
var helpPrefix = regexp.MustCompile(`(?i)^milbot atnd help`)

// redis のクライアント
var redisCli = redis.NewClient(&redis.Options{
	Addr:     "host_address:6379",
	Password: "",
	DB:       0,
})

// KeyMembers はメンバー hash に使う key
const KeyMembers = "atnd_members"

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
			} else if addPrefix.MatchString(ev.Text) {
				go add(api, ev)
			} else if removePrefix.MatchString(ev.Text) {
				go remove(api, ev)
			} else if listPrefix.MatchString(ev.Text) {
				go list(api, ev)
			} else if atndPrefix.MatchString(ev.Text) {
				// これは並列実行できない！
				atnd(api, ev)
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

	mes :=
		"以下の機能があります(｀･ω･´)\n" +
			"`milbot atnd help`\n" +
			"	このメッセージを表示する。\n" +
			"\n" +
			"`milbot atnd`\n" +
			"    出席しているメンバーを確認する。\n" +
			"\n" +
			"`milbot atnd list`\n" +
			"	登録してあるメンバーを確認する。\n" +
			"\n" +
			"`milbot atnd add name bd_addr`\n" +
			"	メンバー登録をする。\n" +
			"	`name` に自分の名前を入れてください。半角スペースは使えません。\n" +
			"	`bd_addr` に Bluetooth アドレスを `00:33:66:99:aa:dd` 形式で書いてください。\n" +
			"\n" +
			"`milbot atnd remove name`\n" +
			"	メンバーを削除します。\n" +
			"	`name` に削除するメンバーの名前を正しく書いてください。\n" +
			"	`milbot atnd list` で正確な名前を確認してください。"

	botutils.SendMessageWithLog(api, ev.Channel, mes)
}

// add でメンバーを追加する
func add(api *slack.Client, ev *slack.MessageEvent) {
	botutils.LogEventReceive(api, ev, "atnd add")

	elems := strings.Split(ev.Text, " ")
	if len(elems) != 5 {
		botutils.SendMessageWithLog(api, ev.Channel, "不正な入力です(´･ω･｀)\nusage: milbot atnd add name address")
		return
	}
	name := elems[3]
	bdAddr := elems[4]
	if !isValidBdAddr(bdAddr) {
		mes := fmt.Sprint("不正な Bluetooth アドレスです(´･ω･｀): ", bdAddr)
		botutils.SendMessageWithLog(api, ev.Channel, mes)
		return
	}

	members, err := redisCli.HKeys(KeyMembers).Result()
	if err != nil {
		botutils.SendMessageWithLog(api, ev.Channel, "データベースにアクセスできません(´; ω ;｀)")
		botutils.LogBoth("atnd add: redis error: ", err)
		return
	}

	for _, member := range members {
		if name == member {
			botutils.SendMessageWithLog(api, ev.Channel, name+" はすでに登録されています(´･ω･｀)")
			return
		}
	}

	if err := redisCli.HSet(KeyMembers, name, bdAddr).Err(); err != nil {
		botutils.SendMessageWithLog(api, ev.Channel, "データベースに登録できませんでした(´; ω ;｀)")
		botutils.LogBoth("atnd add: redis error: ", err)
		return
	}

	mes := fmt.Sprintf("登録しました(｀･ω･´)\n%s: %s", name, bdAddr)
	botutils.SendMessageWithLog(api, ev.Channel, mes)
}

// remove でメンバーを削除する
func remove(api *slack.Client, ev *slack.MessageEvent) {
	botutils.LogEventReceive(api, ev, "atnd remove")

	elems := strings.Split(ev.Text, " ")
	if len(elems) != 4 {
		botutils.SendMessageWithLog(api, ev.Channel, "不正な入力です(´･ω･｀)\nusage: milbot atnd remove name")
		return
	}
	name := elems[3]

	members, err := redisCli.HKeys(KeyMembers).Result()
	if err != nil {
		botutils.SendMessageWithLog(api, ev.Channel, "データベースにアクセスできません(´; ω ;｀)")
		botutils.LogBoth("atnd remove: redis error: ", err)
		return
	}
	exists := false
	for _, member := range members {
		if name == member {
			exists = true
		}
	}
	if !exists {
		botutils.SendMessageWithLog(api, ev.Channel, name+" はもともと登録されていません(´･ω･｀)")
		return
	}

	if err := redisCli.HDel(KeyMembers, name).Err(); err != nil {
		botutils.SendMessageWithLog(api, ev.Channel, "データベースから削除できませんでした(´; ω ;｀)")
		botutils.LogBoth("atnd remove: redis error: ", err)
		return
	}

	botutils.SendMessageWithLog(api, ev.Channel, "削除しました(｀･ω･´)\n"+name)
}

// list でメンバーをリストアップする
func list(api *slack.Client, ev *slack.MessageEvent) {
	botutils.LogEventReceive(api, ev, "atnd list")

	members, err := redisCli.HKeys(KeyMembers).Result()
	if err != nil {
		botutils.SendMessageWithLog(api, ev.Channel, "データベースにアクセスできません(´; ω ;｀)")
		botutils.LogBoth("atnd list: redis error: ", err)
		return
	}

	if len(members) == 0 {
		botutils.SendMessageWithLog(api, ev.Channel, "誰も登録されていません(´･ω･｀)")
		return
	}

	mes := "以下のメンバーが登録されています(｀･ω･´)\n"
	mes += strings.Join(members, "\n")
	botutils.SendMessageWithLog(api, ev.Channel, mes)
}

// atnd でメンバーをサーチする
func atnd(api *slack.Client, ev *slack.MessageEvent) {
	botutils.LogEventReceive(api, ev, "atnd atnd")

	go botutils.SendMessageWithLog(api, ev.Channel, "在室確認をします。しばらくお待ちください(｀･ω･´)")

	memberAddr, err := redisCli.HGetAll(KeyMembers).Result()
	if err != nil {
		go botutils.SendMessageWithLog(api, ev.Channel, "データベースにアクセスできません(´; ω ;｀)")
		go botutils.LogBoth("atnd: redis error: ", err)
		return
	}

	addrMember := swapMap(memberAddr)

	exists, err := postAtnd(memberAddr)
	if err != nil {
		go botutils.SendMessageWithLog(api, ev.Channel, "サーバとの接続に失敗しました(´; ω ;｀)")
		go botutils.LogBoth("atnd: could not access bluetooth server", err)
		return
	}

	if len(exists) == 0 {
		go botutils.SendMessageWithLog(api, ev.Channel, "現在部屋には誰もいません(´･ω･｀)")
		return
	}

	mes := "現在部屋には\n"
	for _, addr := range exists {
		mes += addrMember[addr]
	}
	mes += "が在室しています(｀･ω･´)"
	go botutils.SendMessageWithLog(api, ev.Channel, mes)
}

// Exist は部屋に誰かいたら true を返す
func Exist() bool {
	memberAddr, err := redisCli.HGetAll(KeyMembers).Result()
	if err != nil {
		go botutils.LogBoth("atnd: redis error: ", err)
		return false
	}

	exists, err := postAtnd(memberAddr)
	if err != nil {
		go botutils.LogBoth("atnd: could not access bluetooth server", err)
		return false
	}

	return len(exists) > 0
}

// postAtnd はラズパイのサーバに post して今いるメンバーのアドレスのリストを返す
func postAtnd(memberAddr map[string]string) ([]string, error) {
	url := "http://host_address:" + os.Getenv("ATND_PORT")
	body := ""
	for _, bdAddr := range memberAddr {
		body += bdAddr + "\n"
	}

	resp, err := http.Post(url, "text/plain", strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("server returns %s", resp.Status)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	exists := removeBlanks(strings.Split(string(b), "\n"))
	return exists, nil
}

// isValidBdAddr は bdAddr が有効な bd_addr かどうかを返す
func isValidBdAddr(bdAddr string) bool {
	runes := []rune(bdAddr)
	if len(runes) != 17 {
		return false
	}

	for _, i := range []int{2, 5, 8, 11, 14} {
		if runes[i] != ':' {
			return false
		}
	}

	for _, i := range []int{0, 3, 6, 9, 12, 15} {
		numstr := string(runes[i : i+2])
		if _, err := strconv.ParseUint(numstr, 16, 8); err != nil {
			return false
		}
	}

	return true
}

// swapMap は m の key と value を入れ替えます
func swapMap(m map[string]string) map[string]string {
	ret := make(map[string]string)
	for k, v := range m {
		ret[v] = k
	}
	return ret
}

// getAtndPort は環境変数から atnd サーバのポートを取得する
func getAtndPort() (string, error) {
	port, ok := os.LookupEnv("ATND_PORT")
	if !ok {
		return "", fmt.Errorf("could not find ATND_PORT")
	}
	return port, nil
}

// removeBlanks は strs の中から "" の要素を取り除く
func removeBlanks(strs []string) []string {
	ret := make([]string, len(strs))
	copy(ret, strs)
	for i := 0; i < len(ret); i++ {
		if ret[i] == "" {
			ret[i], ret[len(strs)-1] = ret[len(strs)-1], ret[i]
			ret = ret[:len(strs)-1]
		}
	}
	return ret
}
