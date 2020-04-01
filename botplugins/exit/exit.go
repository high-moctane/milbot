package exit

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/high-moctane/milbot/botlog"
	"github.com/slack-go/slack"
)

// validPrefix は反応するメッセージの正規表現です。
var validRegexp = regexp.MustCompile(`(?i)^milbot exit`)

// Plugin は終了コマンドを受け付けるプラグインです
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

// Serve で終了コマンドを受け付けて終了します。
func (p *Plugin) Serve(ctx context.Context, event slack.RTMEvent) error {
	if !p.isValidEvent(event) {
		return nil
	}

	defer os.Exit(0)

	ev, _ := event.Data.(*slack.MessageEvent)
	user, err := p.getUserNameContext(ctx, ev)
	if err != nil {
		return fmt.Errorf("exit failed: %w", err)
	}

	_, _, _, err = p.client.SendMessageContext(
		ctx,
		ev.Channel,
		slack.MsgOptionText("Bye (｀･ω･´)", true),
	)
	if err != nil {
		return fmt.Errorf("exit failed: %v", err)
	}
	log.Printf("received exit command by %s", user)
	botlog.SendfContext(ctx, "received exit command by %s", user)
	return nil
}

// getUserName は event の送り主の名前を取得します。
func (p *Plugin) getUserNameContext(ctx context.Context, event *slack.MessageEvent) (name string, err error) {
	user, err := p.client.GetUserInfoContext(ctx, event.User)
	if err != nil {
		err = fmt.Errorf("could not get user: %w", err)
		return
	}
	name = user.Name
	return
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
	return "[Exit]\n" +
		"`milbot exit` を受け取って bot を終了します。"
}
