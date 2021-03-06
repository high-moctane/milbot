package atnd

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/high-moctane/milbot/libatnd"
	"github.com/slack-go/slack"
)

// 反応する regexp たちです。
var regexpAtndSet = regexp.MustCompile(`(?i)^milbot atnd set`)
var regexpAtndDelete = regexp.MustCompile(`(?i)^milbot atnd delete`)
var regexpAtndList = regexp.MustCompile(`(?i)^milbot atnd list`)
var regexpAtnd = regexp.MustCompile(`(?i)^milbot atnd`)

// Plugin は 在室状況を確認するプラグインです。
type Plugin struct {
	client *slack.Client
	atnd   *libatnd.Atnd
}

// New でプラグインを生成します。
func New() *Plugin {
	return new(Plugin)
}

// Start でプラグインを有効化します。
func (p *Plugin) Start(client *slack.Client) error {
	p.client = client
	p.atnd = libatnd.Instance()
	return nil
}

// Serve で atnd に反応してメッセージを返します。
func (p *Plugin) Serve(ctx context.Context, event slack.RTMEvent) error {
	ev, ok := event.Data.(*slack.MessageEvent)
	if !ok {
		return nil
	}

	if ev.BotID != "" {
		return nil
	}

	if p.isAtndSetQuery(ev) {
		if err := p.serveAtndSet(ctx, ev); err != nil {
			return fmt.Errorf("atnd serve error: %w", err)
		}
	} else if p.isAtndDeleteQuery(ev) {
		if err := p.serveAtndDelete(ctx, ev); err != nil {
			return fmt.Errorf("atnd serve error: %w", err)
		}
	} else if p.isAtndListQuery(ev) {
		if err := p.serveAtndList(ctx, ev); err != nil {
			return fmt.Errorf("atnd serve error: %w", err)
		}
	} else if p.isAtndQuery(ev) {
		if err := p.serveAtnd(ctx, ev); err != nil {
			return fmt.Errorf("atnd serve error: %w", err)
		}
	}

	return nil
}

func (*Plugin) isAtndSetQuery(ev *slack.MessageEvent) bool {
	return regexpAtndSet.MatchString(ev.Text)
}

func (p *Plugin) serveAtndSet(ctx context.Context, event *slack.MessageEvent) error {
	elems := strings.Split(event.Text, " ")
	if len(elems) != 5 {
		_, _, _, err := p.client.SendMessageContext(
			ctx,
			event.Channel,
			slack.MsgOptionText("フォーマットが違います。`milbot help` をご覧ください (´･ω･｀)", true),
		)
		if err != nil {
			return fmt.Errorf("serve atnd set error: %w", err)
		}
		return nil
	}

	name, addr := elems[3], elems[4]
	err := p.atnd.SetMember(name, addr)
	var macErr libatnd.InvalidMACAddressError
	var nameErr libatnd.InvalidNameError
	if errors.As(err, &macErr) {
		_, _, _, err := p.client.SendMessageContext(
			ctx,
			event.Channel,
			slack.MsgOptionText("変な Bluetooth アドレスです (´･ω･｀)", true),
		)
		if err != nil {
			return fmt.Errorf("serve atnd set error: %w", err)
		}
		return nil
	} else if errors.As(err, &nameErr) {
		_, _, _, err := p.client.SendMessageContext(
			ctx,
			event.Channel,
			slack.MsgOptionText("その名前は使えません (´･ω･｀)", true),
		)
		if err != nil {
			return fmt.Errorf("serve atnd set error: %w", err)
		}
		return nil
	} else if err != nil {
		return fmt.Errorf("serve atnd set error: %w", err)
	}

	_, _, _, err = p.client.SendMessageContext(
		ctx,
		event.Channel,
		slack.MsgOptionText("登録しました (｀･ω･´)", true),
	)
	if err != nil {
		return fmt.Errorf("serve atnd set error: %w", err)
	}

	return nil
}

func (p *Plugin) isAtndDeleteQuery(ev *slack.MessageEvent) bool {
	return regexpAtndDelete.MatchString(ev.Text)
}

func (p *Plugin) serveAtndDelete(ctx context.Context, event *slack.MessageEvent) error {
	elems := strings.Split(event.Text, " ")
	if len(elems) != 4 {
		_, _, _, err := p.client.SendMessageContext(
			ctx,
			event.Channel,
			slack.MsgOptionText("フォーマットが違います。`milbot help` をご覧ください (´･ω･｀)", true),
		)
		if err != nil {
			return fmt.Errorf("serve atnd delete error: %w", err)
		}
		return nil
	}

	name := elems[3]
	err := p.atnd.DeleteMember(name)
	var notExistErr libatnd.MemberNotExistError
	if errors.As(err, &notExistErr) {
		_, _, _, err := p.client.SendMessageContext(
			ctx,
			event.Channel,
			slack.MsgOptionText("その名前のメンバーはいません (´･ω･｀)", true),
		)
		if err != nil {
			return fmt.Errorf("serve atnd delete error: %w", err)
		}
		return nil
	} else if err != nil {
		return fmt.Errorf("serve atnd delete error: %w", err)
	}

	_, _, _, err = p.client.SendMessageContext(
		ctx,
		event.Channel,
		slack.MsgOptionText("削除しました (｀･ω･´)", true),
	)
	if err != nil {
		return fmt.Errorf("serve atnd delete error: %w", err)
	}

	return nil
}

func (p *Plugin) isAtndListQuery(ev *slack.MessageEvent) bool {
	return regexpAtndList.MatchString(ev.Text)
}

func (p *Plugin) serveAtndList(ctx context.Context, event *slack.MessageEvent) error {
	list := p.atnd.Members()

	msg := new(strings.Builder)

	msg.WriteString("現在\n")

	for _, name := range list {
		msg.WriteString(name)
		msg.WriteString("\n")
	}

	msg.WriteString("が登録されています (｀･ω･´)")

	_, _, _, err := p.client.SendMessageContext(
		ctx,
		event.Channel,
		slack.MsgOptionText(msg.String(), true),
	)
	if err != nil {
		return fmt.Errorf("serve atnd list error: %w", err)
	}

	return nil
}

func (p *Plugin) isAtndQuery(ev *slack.MessageEvent) bool {
	return regexpAtnd.MatchString(ev.Text)
}

func (p *Plugin) serveAtnd(ctx context.Context, event *slack.MessageEvent) error {
	if err := p.sendHistoryMessage(ctx, event.Channel); err != nil {
		return fmt.Errorf("serve atnd error: %w", err)
	}

	// あらためて在室確認する。
	if err := p.sendAttendanceMessage(ctx, event.Channel); err != nil {
		return fmt.Errorf("serve atnd error: %w", err)
	}

	return nil
}

// sendHistoryMessage でこれまでの在室履歴を送信します。
func (p *Plugin) sendHistoryMessage(ctx context.Context, channel string) error {
	history := p.atnd.Status()
	_, _, _, err := p.client.SendMessageContext(
		ctx,
		channel,
		slack.MsgOptionText(p.historyMessage(history), true),
	)
	if err != nil {
		return fmt.Errorf("send history message failed: %w", err)
	}
	return nil
}

// historyMessage は在室履歴のメッセージを構築します。
func (p *Plugin) historyMessage(history []*libatnd.Attendance) string {
	if len(history) == 0 {
		return "Bot を起動してから誰も在室していないようです (´･ω･｀)"
	}

	msg := new(strings.Builder)
	msg.WriteString("これまでの在室履歴は\n")
	now := time.Now()
	for _, mem := range history {
		msg.WriteString(fmt.Sprintf("%s: %s\n", mem.Name, p.timeDiffFormat(now, mem.Time)))
	}
	msg.WriteString("です (｀･ω･´)")

	return msg.String()
}

// timeDiffFormat はt と now の差をわかりやすいフォーマットに変換します。
func (*Plugin) timeDiffFormat(now, t time.Time) string {
	duration := now.Sub(t)

	if duration.Hours() >= 24.0 {
		return fmt.Sprintf("%.0f 日前", duration.Hours()/24.0)
	} else if duration.Hours() >= 1.0 {
		return fmt.Sprintf("%.0f 時間前", duration.Hours())
	} else if duration.Minutes() >= 1.0 {
		return fmt.Sprintf("%.0f 分前", duration.Minutes())
	}
	return fmt.Sprintf("%.0f 秒前", duration.Seconds())
}

// sendAttendanceMessage は現在の在室状況を送信します。
func (p *Plugin) sendAttendanceMessage(ctx context.Context, channel string) error {
	_, _, _, err := p.client.SendMessageContext(
		ctx,
		channel,
		slack.MsgOptionText("在室確認します。しばらくお待ちください……(｀･ω･´)", true),
	)
	if err != nil {
		return fmt.Errorf("send attendance message failed: %w", err)
	}

	attendance, err := p.atnd.SearchContext(ctx)
	if errors.Is(err, libatnd.ErrBluetoothNotAvailable) {
		_, _, _, err := p.client.SendMessageContext(
			ctx,
			channel,
			slack.MsgOptionText("Bluetooth が死んでます (´; ω ;｀)", true),
		)
		if err != nil {
			return fmt.Errorf("send attendance message failed: %w", err)
		}
		return nil
	} else if err != nil {
		return fmt.Errorf("send attendance message failed: %w", err)
	}

	_, _, _, err = p.client.SendMessageContext(
		ctx,
		channel,
		slack.MsgOptionText(p.attendanceMessage(attendance), true),
	)
	if err != nil {
		return fmt.Errorf("send attendance message failed: %w", err)
	}

	return nil
}

// attendanceMessage は出席している人のメッセージを返します。
func (*Plugin) attendanceMessage(attendance []*libatnd.Attendance) string {
	if len(attendance) == 0 {
		return "現在研究室には誰もいません (´･ω･｀)"
	}

	msg := new(strings.Builder)
	msg.WriteString("現在研究室には\n")
	for _, mem := range attendance {
		msg.WriteString(mem.Name)
		msg.WriteString("\n")
	}
	msg.WriteString("が在室しています (｀･ω･´)")

	return msg.String()
}

// Stop でプラグインの終了処理をします。
func (p *Plugin) Stop() error {
	return nil
}

// Help でヘルプメッセージを返します。
func (p *Plugin) Help() string {
	return "[Atnd]\n" +
		"研究室に在室しているメンバーを調べます (｀･ω･´)\n" +
		"\n" +
		"`milbot atnd`:\n" +
		"現在の在室状況をお知らせします。\n" +
		"\n" +
		"`milbot atnd set <name> <bluetooth address>`\n" +
		"メンバー登録または変更をします。\n" +
		"`<name>` に自分の名前，`<bluetooth address>` に自分のスマートフォンの Bluetooth アドレスを入力してください。\n" +
		"例: `milbot atnd set 俺様 12:34:56:78:90:ab`\n" +
		"\n" +
		"`milbot atnd delete <name>`\n" +
		"メンバーを削除します。\n" +
		"<name> に自分の名前をいれてください。\n" +
		"\n" +
		"例: `milbot atnd delete 俺様`" +
		"`milbot atnd list`\n" +
		"登録されているメンバーの名前を表示します。"
}
