package main

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/high-moctane/milbot/botlog"
	"github.com/high-moctane/milbot/botplugin"
	"github.com/high-moctane/milbot/botplugins/ping"
	_ "github.com/joho/godotenv/autoload"
	"github.com/slack-go/slack"
)

// pluginTimeout はプラグインが返事をするののタイムアウト時間です。
var pluginTimeout = 120 * time.Second

// plugins にプラグインを入れていくぞ(｀･ω･´)！
var plugins = []botplugin.Plugin{
	ping.New(),
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
	// 起動ログ
	log.Print("milbot launch(｀･ω･´)！")

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

	// 受け取ったイベントをプラグインに渡していく
	ctx := context.Background()
	for event := range rtm.IncomingEvents {
		if err := detectUncontinuableRTMEvent(&event); err != nil {
			return err
		}

		if isExitMessage(event) {
			log.Printf("received exit message")
			botlog.Sendf("received exit message")
			return nil
		}

		for _, plg := range append(plugins, helpPlugin) {
			ctx, cancel := context.WithTimeout(ctx, pluginTimeout)
			go func(plg botplugin.Plugin) {
				if err := plg.Serve(ctx, client, event); err != nil {
					log.Print(err)
				}
				cancel()
			}(plg)
		}
	}

	return nil
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
