package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/high-moctane/milbot/milbot/plugin"
	"github.com/high-moctane/milbot/milbot/plugins/exit"
	"github.com/high-moctane/milbot/milbot/plugins/ping"
	_ "github.com/joho/godotenv/autoload"
	"github.com/nlopes/slack"
)

// logger はちょっとリッチにしといた
var logger = log.New(os.Stdout, "milbot: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)

// ここにプラグインを列挙していくぞ！
var plugins = []plugin.Plugin{
	ping.New(),
	exit.New(),
}

func main() {
	if err := run(); err != nil {
		logger.Fatal(err)
	}
}

func run() error {
	go handleSignal()

	api, err := newAPI()
	if err != nil {
		return err
	}
	rtm := api.NewRTM()

	go rtm.ManageConnection()
	defer rtm.Disconnect()

	// この chan を使うことで Plugin が死んだかどうか
	// つまり rtm.IncomingEvents が閉じたかどうかを確認する
	ch := make(chan struct{}, len(plugins))
	for _, p := range plugins {
		ch <- struct{}{}
		go func(p plugin.Plugin) {
			appendPlugin(p, api, rtm.IncomingEvents)
			<-ch
		}(p) // このへんなのは golang の仕様
	}

	// ここがブロックされないということはどれかの Plugin が死んだ。
	// どれかのプラグインが死んだら即 bot 終了という仕様にする
	ch <- struct{}{}
	return fmt.Errorf("main loop ended")
}

// appendPlugin を go で呼び出すとプラグインが走り出す
func appendPlugin(p plugin.Plugin, api *slack.Client, eventCh chan slack.RTMEvent) {
	p.Serve(api, eventCh)
	defer p.Stop()
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
