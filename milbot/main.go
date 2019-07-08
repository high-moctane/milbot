package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/high-moctane/milbot/milbot/plugin"
	"github.com/high-moctane/milbot/milbot/plugins/atnd"
	"github.com/high-moctane/milbot/milbot/plugins/exit"
	"github.com/high-moctane/milbot/milbot/plugins/hello"
	"github.com/high-moctane/milbot/milbot/plugins/peng"
	"github.com/high-moctane/milbot/milbot/plugins/ping"
	"github.com/high-moctane/milbot/milbot/plugins/restart"
	_ "github.com/joho/godotenv/autoload"
	"github.com/nlopes/slack"
)

// logger はちょっとリッチにしといた
var logger = log.New(os.Stdout, "milbot: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)

// ここにプラグインを列挙していくぞ！
var plugins = []plugin.Plugin{
	atnd.New(),
	exit.New(),
	hello.New(),
	peng.New(),
	ping.New(),
	restart.New(),
}

func main() {
	if err := run(); err != nil {
		logger.Fatal(err)
	}
}

func run() error {
	// シグナルハンドリング
	go handleSignal()

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
	for i := 0; i < len(plugins); i++ {
		go appendPlugin(plugins[i], api, eventChs[i])
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

	// ここにたどり着いたということは rtm.IncomingEvents がもう受け取れないということ
	return fmt.Errorf("main loop ended")
}

// appendPlugin を go で呼び出すとプラグインが走り出す
func appendPlugin(p plugin.Plugin, api *slack.Client, eventCh chan slack.RTMEvent) {
	p.Serve(api, eventCh)
	defer p.Stop()
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
		slack.OptionDebug(true), // これを true にすると通信の詳細が表示される
		slack.OptionLog(logger),
	), nil
}

// handleSignal でシグナルをハンドリングします
func handleSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	s := <-ch
	switch s {
	case syscall.SIGTERM:
		logger.Print("caught SIGTERM")
		os.Exit(1)

	case syscall.SIGINT:
		logger.Print("caught SIGINT")
		os.Exit(1)

	default:
		logger.Printf("caught unknown signal: %s", s)
	}
}
