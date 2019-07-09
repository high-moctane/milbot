package script

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
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

	mes := "スクリプトを走らせます。\n" +
		"現在以下のコマンドを受け付けています。\n" +
		"    `milbot bash`\n" +
		"    `milbot python3`\n" +
		"コマンドに続けてスクリプトを入力してください。" +
		"コードを ``` で囲んでも動きます。"

	botutils.SendMessageWithLog(api, ev, mes)
}

// bash を走らせる
func bash(api *slack.Client, ev *slack.MessageEvent) {
	input, err := parseInput(ev.Text)
	if err != nil {
		botutils.SendMessageWithLog(api, ev, "入力フォーマットが不正です(´･ω･｀)")
		return
	}
	out, err := run([]string{"bash"}, "main.sh", input)
	if err != nil {
		botutils.SendMessageWithLog(api, ev, "実行に失敗しました(´; ω ;｀)\n"+out)
		botutils.LogBoth("script: bash error: ", err)
		return
	}
	botutils.SendMessageWithLog(api, ev, out)
}

// python3 を走らせる
func python3(api *slack.Client, ev *slack.MessageEvent) {
	input, err := parseInput(ev.Text)
	if err != nil {
		botutils.SendMessageWithLog(api, ev, "入力フォーマットが不正です(´･ω･｀)")
		return
	}
	out, err := run([]string{"python3", "-B"}, "main.py", input)
	if err != nil {
		botutils.SendMessageWithLog(api, ev, "実行に失敗しました(´; ω ;｀)\n"+out)
		botutils.LogBoth("script: python3 error: ", err)
		return
	}
	botutils.SendMessageWithLog(api, ev, out)
}

// parseInput は送られたメッセージからスクリプトの部分を取り出す
func parseInput(text string) (string, error) {
	// TODO: regexp を関数の外で定義する
	r := regexp.MustCompile(`\s.`)
	indices := r.FindAllIndex([]byte(text), 2)
	if len(indices) < 2 {
		return "", fmt.Errorf("not found body")
	}

	body := text[indices[1][1]-1:]
	body = strings.TrimPrefix(body, "```")
	body = strings.TrimSuffix(body, "```")
	return body, nil
}

// run で実際にコマンドを走らせる
func run(cmd []string, fname, script string) (string, error) {
	dir, err := ioutil.TempDir("", "milbot_script")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)

	abspath := filepath.Join(dir, fname)
	if err := writeScript(abspath, script); err != nil {
		return "", err
	}

	var c *exec.Cmd
	if len(cmd) == 1 {
		c = exec.Command(cmd[0], abspath)
	} else {
		c = exec.Command(cmd[0], append(cmd[1:], abspath)...)
	}
	out, err := c.CombinedOutput()
	if err != nil {
		return string(out), err
	}
	return string(out), err
}

// writeScript で abspath に script が書かれたファイルを作ります
func writeScript(abspath, script string) error {
	f, err := os.Create(abspath)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(script); err != nil {
		return err
	}
	return nil
}
