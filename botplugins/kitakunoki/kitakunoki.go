package kitakunoki

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/high-moctane/milbot/libatnd"
	"github.com/robfig/cron/v3"
	"github.com/slack-go/slack"
)

// cronScuedule はいつ帰宅の木をするかのスケジュールです。
const cronSchedule = "* 21 * * *"

// Plugin は帰宅を促します。
type Plugin struct {
	client       *slack.Client
	kitakunoList []*kitakunoEntry
	cron         *cron.Cron
	atnd         *libatnd.Atnd
}

// New でプラグインを生成します。
func New() *Plugin {
	return new(Plugin)
}

// Start でプラグインを有効化します。
func (p *Plugin) Start(client *slack.Client) error {
	p.client = client

	kitakunoList, err := kitakunoList()
	if err != nil {
		return fmt.Errorf("kitakunoki start failed: %w", err)
	}
	p.kitakunoList = kitakunoList

	p.atnd = libatnd.Instance()

	p.cron = cron.New()
	p.cron.AddFunc(cronSchedule, func() {
		if err := p.kitakunoDo(); err != nil {
			log.Print(err)
		}
	})
	p.cron.Start()

	return nil
}

// kitakunoDo は研究室に人がいる場合に kitakunoPost します。
func (p *Plugin) kitakunoDo() error {
	attendance, err := p.atnd.Search()
	if err != nil {
		return fmt.Errorf("kitakuno do failed: %w", err)
	}
	if len(attendance) == 0 {
		return nil
	}

	if err := p.kitakunoPost(); err != nil {
		return fmt.Errorf("kitakuno do failed: %w", err)
	}

	return nil
}

// kitakunoPost は帰宅の木をお知らせします。
func (p *Plugin) kitakunoPost() error {
	ki := p.randomKitakunoki()
	_, _, _, err := p.client.SendMessage("#random", slack.MsgOptionText(
		p.kitakunoMessage(ki), true,
	))
	if err != nil {
		return fmt.Errorf("kitakuno post error: %w", err)
	}
	return nil
}

// kitakunoMessage は帰宅の木のメッセージを構築します。
func (p *Plugin) kitakunoMessage(ki *kitakunoEntry) string {
	return fmt.Sprintf("今日の帰宅の木は\n%s\n%s\nです (｀･ω･´):evergreen_tree:",
		ki.name, ki.url)
}

// randomKitakunoki はランダムな帰宅の木を返します。
func (p *Plugin) randomKitakunoki() *kitakunoEntry {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	idx := rnd.Intn(len(p.kitakunoList))
	return p.kitakunoList[idx]
}

// Serve はとくに何もしません。
func (p *Plugin) Serve(_ context.Context, _ slack.RTMEvent) error {
	return nil
}

// Stop はとくになにもしません。
func (p *Plugin) Stop() error {
	return nil
}

// Help でヘルプメッセージを返します。
func (p *Plugin) Help() string {
	return "[Kitakunoki]\n" +
		"毎日 21:00 に帰宅を促します (｀･ω･´)"
}
