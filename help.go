package main

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/high-moctane/milbot/botplugin"
	"github.com/slack-go/slack"
)

// HelpPlugin はヘルプメッセージを返すプラグインです
type HelpPlugin struct {
	client      *slack.Client
	plugins     []botplugin.Plugin
	validRegexp *regexp.Regexp
}

// NewHelpPlugin でプラグインを生成します。plugins にプラグインリストを与えます。
func NewHelpPlugin(plugins []botplugin.Plugin) *HelpPlugin {
	return &HelpPlugin{
		plugins:     plugins,
		validRegexp: regexp.MustCompile(`(?i)^mil help`),
	}
}

// Start でプラグインを有効化します。
func (p *HelpPlugin) Start(client *slack.Client) error {
	p.client = client
	return nil
}

// Serve でヘルプメッセージを返します。
func (p *HelpPlugin) Serve(ctx context.Context, event slack.RTMEvent) error {
	if !p.isValidEvent(event) {
		return nil
	}

	ev := event.Data.(*slack.MessageEvent)
	_, _, _, err := p.client.SendMessageContext(
		ctx,
		ev.Channel,
		slack.MsgOptionText(p.buildHelpMessage(), true),
	)
	if err != nil {
		return fmt.Errorf("help failed: %w", err)
	}
	return nil
}

// isValidEvent は event に反応するべきかどうか返します。
func (p *HelpPlugin) isValidEvent(event slack.RTMEvent) bool {
	ev, ok := event.Data.(*slack.MessageEvent)
	if !ok {
		return false
	}
	return p.validRegexp.MatchString(ev.Text)
}

// buildHelpMessage は plugins からヘルプメッセージを生成します。
func (p *HelpPlugin) buildHelpMessage() string {
	helps := []string{}
	for _, plg := range append(p.plugins, p) {
		helps = append(helps, plg.Help())
	}
	return strings.Join(helps, "\n\n")
}

// Stop でプラグインの終了処理をします。
func (p *HelpPlugin) Stop() error {
	return nil
}

// Help でヘルプメッセージを返します。
func (p *HelpPlugin) Help() string {
	return "[Help]\n" +
		"`milbot help` でこのメッセージを表示します。"
}
