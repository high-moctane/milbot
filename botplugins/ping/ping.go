package ping

import (
	"context"
	"fmt"
	"regexp"

	"github.com/slack-go/slack"
)

// validPrefix は反応するメッセージの正規表現です。
var validRegexp = regexp.MustCompile(`(?i)^mil ping`)

// Plugin は ping に pong するプラグインです。
type Plugin struct {
	client *slack.Client
}

// New でプラグインを生成します。
func New() *Plugin {
	return new(Plugin)
}

// Start でプラグインを有効化します。
func (p *Plugin) Start(client *slack.Client) error {
	p.client = client
	return nil
}

// Serve で ping に対して pong を返します。
func (p *Plugin) Serve(ctx context.Context, event slack.RTMEvent) error {
	if !p.isValidEvent(event) {
		return nil
	}

	ev := event.Data.(*slack.MessageEvent)
	_, _, _, err := p.client.SendMessageContext(
		ctx,
		ev.Channel,
		slack.MsgOptionText("pong(｀･ω･´)", true),
	)
	if err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}
	return nil
}

// isValidEvent は event に反応するべきかどうか返します。
func (*Plugin) isValidEvent(event slack.RTMEvent) bool {
	ev, ok := event.Data.(*slack.MessageEvent)
	if !ok {
		return false
	}
	return validRegexp.MatchString(ev.Text)
}

// Stop でプラグインの終了処理をします。
func (p *Plugin) Stop() error {
	return nil
}

// Help でヘルプメッセージを返します。
func (p *Plugin) Help() string {
	return "[Ping]\n" +
		"`milbot ping` に pong を返します。\n" +
		"Bot の生存確認に使ってください(｀･ω･´)"
}
