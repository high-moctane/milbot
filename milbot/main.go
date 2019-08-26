package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/high-moctane/milbot/milbot/botutils"
	"github.com/high-moctane/milbot/milbot/plugin"
	"github.com/high-moctane/milbot/milbot/plugins/atnd"
	"github.com/high-moctane/milbot/milbot/plugins/exit"
	"github.com/high-moctane/milbot/milbot/plugins/hello"
	"github.com/high-moctane/milbot/milbot/plugins/help"
	"github.com/high-moctane/milbot/milbot/plugins/kitakunoki"
	"github.com/high-moctane/milbot/milbot/plugins/peng"
	"github.com/high-moctane/milbot/milbot/plugins/ping"
	"github.com/high-moctane/milbot/milbot/plugins/restart"
	"github.com/high-moctane/milbot/milbot/plugins/script"
	"github.com/high-moctane/milbot/milbot/plugins/verse"
	_ "github.com/joho/godotenv/autoload"
	"github.com/nlopes/slack"
)

// ここにプラグインを列挙していくぞ！
var plugins = []plugin.Server{
	atnd.New(),
	exit.New(),
	hello.New(),
	help.New(),
	kitakunoki.New(),
	peng.New(),
	ping.New(),
	restart.New(),
	script.New(),
	verse.New(),
}

func main() {
	if err := run(); err != nil {
		botutils.LogBoth("main error: ", err)
	}
}

func run() error {
	// Slack に接続
	api, err := newAPI()
	if err != nil {
		return err
	}
	rtm := api.NewRTM()

	go rtm.ManageConnection()
	defer rtm.Disconnect()

	// 各プラグインに与えるイベント chan
	eventChs := makeEventChs()

	// プラグインを起動していく
	ctx, cancel := context.WithCancel(context.Background())
	wg := new(sync.WaitGroup)
	for i := 0; i < len(plugins); i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			appendPlugin(ctx, plugins[i], api, eventChs[i])
		}(i)
	}

	// ここでイベントを受け取り各プラグインに一斉送信する
	for ev := range rtm.IncomingEvents {
		for i := 0; i < len(plugins); i++ {
			// こういう実装は goroutine leak を招くが，
			// 実際そんなにやばいリクエストは来ない
			go func(i int) {
				eventChs[i] <- ev
			}(i)
		}
	}

	// signal を受け取って run を終了する
	s := <-signalChan()
	switch s {
	case syscall.SIGTERM:
		botutils.LogBoth("caught SIGTERM")

	case syscall.SIGINT:
		botutils.LogBoth("caught SIGINT")

	default:
		botutils.LogBoth("caught unknown signal: ", s)
	}
	cancel()
	wg.Wait()
	return fmt.Errorf("exit run")
}

// appendPlugin を go で呼び出すとプラグインが走り出す
func appendPlugin(ctx context.Context, p plugin.Server, api *slack.Client, eventCh chan slack.RTMEvent) {
	p.Serve(ctx, api, eventCh)
}

// makeEventChs は各プラグインに与える chan の列を生成する
func makeEventChs() []chan slack.RTMEvent {
	chs := make([]chan slack.RTMEvent, len(plugins))
	for i := 0; i < len(plugins); i++ {
		// バッファが 10 もあれば十分でしょ！ 10 だけに(｀･ω･´)
		chs[i] = make(chan slack.RTMEvent, 10)
	}
	return chs
}

// getSlackToken で SLACK_API_TOKEN を取得する
func getSlackToken() (string, error) {
	token, ok := os.LookupEnv("SLACK_API_TOKEN")
	if !ok {
		return "", fmt.Errorf("slack api token not found")
	}
	return token, nil
}

// newAPI は *slack.Client を取得する
func newAPI() (*slack.Client, error) {
	token, err := getSlackToken()
	if err != nil {
		return nil, err
	}

	return slack.New(
		token,
		slack.OptionDebug(false), // これを true にすると通信の詳細が表示される
	), nil
}

// handleSignal でシグナルをハンドリングします
func signalChan() <-chan os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	return ch
}
