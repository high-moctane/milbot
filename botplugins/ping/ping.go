package ping

import (
	"context"
	"regexp"

	"github.com/slack-go/slack"
)

// validPrefix は反応するメッセージの正規表現です。
var validRegexp = regexp.MustCompile(`(?i)^mil ping`)

// Plugin は ping に pong するプラグインです。
type Plugin struct{}

// New でプラグインを生成します。
func New() *Plugin {
	return new(Plugin)
}

// Start でプラグインを有効化します。
func (p *Plugin) Start() error {
	return nil
}

// Serve で ping に対して pong を返します。
func (p *Plugin) Serve(ctx context.Context, client *slack.Client, event slack.RTMEvent) error {
	if !p.isValidEvent(event) {
		return nil
	}

	ev := event.Data.(*slack.MessageEvent)
	_, _, _, err := client.SendMessageContext(
		ctx,
		ev.Channel,
		slack.MsgOptionText("pong(｀･ω･´)", true),
	)
	return err
}

// isValidEvent は event に反応するべきかどうか返します。
func (*Plugin) isValidEvent(event slack.RTMEvent) bool {
	ev, ok := event.Data.(*slack.MessageEvent)
	if !ok {
		return false
	}
	return validRegexp.MatchString(ev.Text)
}

// Help でヘルプメッセージを返します。
func (p *Plugin) Help() string {
	return "## Ping\n" +
		"`milbot ping` に pong を返します。\n" +
		"Bot の生存確認に使ってください(｀･ω･´)"
}

// Stop でプラグインの終了処理をします。
func (p *Plugin) Stop() error {
	return nil
}
