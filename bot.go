package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/high-moctane/milbot/botplugin"
	"github.com/slack-go/slack"
)

// envSlackClientSecret は Slack Client Secret の環境変数です。
const envSlackClientSecret = "MILBOT_SLACK_CLIENT_SECRET"

// pluginTimeout はプラグインが返事をするののタイムアウト時間です。
var pluginTimeout = 120 * time.Second

// Bot は milbot の bot 部分を扱います。
// 終わるときは必ず Stop を呼んでください。
type Bot struct {
	plugins   []botplugin.Plugin
	client    *slack.Client
	rtm       *slack.RTM
	isStarted bool
}

// NewBot は新しい Bot インスタンスを返します。
func NewBot(plugins []botplugin.Plugin) *Bot {
	return &Bot{
		plugins: append(plugins, NewHelpPlugin(plugins)),
	}
}

// Serve は Bot プラグインを起動します。err は 必ず not-nil です。
func (b *Bot) Serve(ctx context.Context) error {
	client, err := b.launchSlack()
	if err != nil {
		return fmt.Errorf("bot run failed: %w", err)
	}
	b.client = client

	if err := b.startPlugins(); err != nil {
		return fmt.Errorf("bot run failed: %w", err)
	}
	if err := b.servePlugins(ctx); err != nil {
		return fmt.Errorf("bot run failed: %w", err)
	}
	return nil
}

// launchSlack で Slack のクライアントを起動します。
func (b *Bot) launchSlack() (client *slack.Client, err error) {
	client, err = b.newSlackClient()
	if err != nil {
		err = fmt.Errorf("launch slack failed: %w", err)
		return
	}
	b.rtm = client.NewRTM()
	go b.rtm.ManageConnection()
	return
}

// newSlackClient で Slack の client を作ります。
func (b *Bot) newSlackClient() (*slack.Client, error) {
	token, err := b.getSlackClientSecret()
	if err != nil {
		return nil, err
	}

	client := slack.New(token, slack.OptionDebug(false))
	return client, err
}

// startPlugin で plugins の起動処理をします。
func (b *Bot) startPlugins() error {
	for _, plg := range b.plugins {
		if err := plg.Start(b.client); err != nil {
			return fmt.Errorf("plugin start failed: %w", err)
		}
	}
	return nil
}

// servePlugins で plugin がそれぞれイベントを受け取ります。
func (b *Bot) servePlugins(ctx context.Context) error {
	for event := range b.rtm.IncomingEvents {
		if err := b.detectUncontinuableRTMEvent(&event); err != nil {
			return fmt.Errorf("serve plugins error, %w", err)
		}

		for _, plg := range b.plugins {
			go b.sendEventToPlugin(ctx, plg, event)
		}
	}
	return nil
}

// detectErrorEvent Bot を終了すべき RTMEvent を見つけます。
func (*Bot) detectUncontinuableRTMEvent(event *slack.RTMEvent) error {
	// TODO いろんなエラーに対応したい
	if _, ok := event.Data.(*slack.InvalidAuthEvent); ok {
		err := errors.New("invalid auth")
		return err
	}
	return nil
}

// sendEventToPlugin は plugin に event を渡します。
func (*Bot) sendEventToPlugin(ctx context.Context, plg botplugin.Plugin, event slack.RTMEvent) {
	newCtx, cancel := context.WithTimeout(ctx, pluginTimeout)
	defer cancel()
	if err := plg.Serve(newCtx, event); err != nil {
		log.Print(err)
	}
}

// getSlackClientSecret は環境変数から Slack API token を取得します。
func (*Bot) getSlackClientSecret() (string, error) {
	token, ok := os.LookupEnv(envSlackClientSecret)
	if !ok {
		return "", errors.New(envSlackClientSecret + " not found")
	}
	return token, nil
}

// Stop は Bot の終了処理をします。必ず呼んでください。
func (b *Bot) Stop() []error {
	var errs []error
	if err := b.rtm.Disconnect(); err != nil {
		errs = append(errs, err)
	}

	if b.isStarted {
		for _, plg := range b.plugins {
			if err := plg.Stop(); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errs
}
