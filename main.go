package main

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/high-moctane/milbot/botlog"
	"github.com/high-moctane/milbot/botplugin"
	"github.com/high-moctane/milbot/botplugins/atnd"
	"github.com/high-moctane/milbot/botplugins/exit"
	"github.com/high-moctane/milbot/botplugins/ping"
	_ "github.com/joho/godotenv/autoload"
	"github.com/slack-go/slack"
)

// pluginTimeout はプラグインが返事をするののタイムアウト時間です。
var pluginTimeout = 120 * time.Second

// plugins にプラグインを入れていくぞ(｀･ω･´)！
var plugins = []botplugin.Plugin{
	ping.New(),
	exit.New(),
	atnd.New(),
}

func main() {
	if err := run(); err != nil {
		botlog.Sendf("milbot terminated with non-zero status code: %v", err)
		log.Fatal(err)
	}
}

// run は実質の main 関数です。err != nil のときに 0 でない終了コードで
// プログラムを終えます。
func run() error {
	// ログ
	log.Print("milbot launch(｀･ω･´)！")
	defer log.Print("milbot exited(｀･ω･´)")

	// プラグインの起動
	for _, plg := range plugins {
		if err := plg.Start(); err != nil {
			botlog.Send(err)
			log.Print(err)
			return err
		}
		defer func(plg botplugin.Plugin) {
			if err := plg.Stop(); err != nil {
				botlog.Send(err)
				log.Print(err)
			}
		}(plg)
	}
	helpPlugin := NewHelpPlugin(plugins)

	// Slack の準備
	client, err := newSlackClient()
	if err != nil {
		return err
	}
	rtm := client.NewRTM()
	go rtm.ManageConnection()
	defer rtm.Disconnect()

	// 受け取ったイベントを処理する
	ctx := context.Background()
	for event := range rtm.IncomingEvents {
		if err := detectUncontinuableRTMEvent(&event); err != nil {
			return err
		}

		for _, plg := range append(plugins, helpPlugin) {
			go sendEventToPlugin(ctx, plg, client, event)
		}
	}

	return nil
}

// sendEventToPlugin は plugin に event を渡します。
func sendEventToPlugin(ctx context.Context, plg botplugin.Plugin, client *slack.Client, event slack.RTMEvent) {
	newCtx, cancel := context.WithTimeout(ctx, pluginTimeout)
	defer cancel()
	if err := plg.Serve(newCtx, client, event); err != nil {
		log.Print(err)
	}
}

// detectErrorEvent Bot を終了すべき RTMEvent を見つけます。
func detectUncontinuableRTMEvent(event *slack.RTMEvent) error {
	// TODO いろんなエラーに対応したい
	if _, ok := event.Data.(*slack.InvalidAuthEvent); ok {
		err := errors.New("invalid auth")
		return err
	}
	return nil
}
