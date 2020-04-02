package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/high-moctane/milbot/botlog"
	"github.com/high-moctane/milbot/botplugin"
	"github.com/high-moctane/milbot/botplugins/atnd"
	"github.com/high-moctane/milbot/botplugins/exit"
	"github.com/high-moctane/milbot/botplugins/kitakunoki"
	"github.com/high-moctane/milbot/botplugins/ping"
	"github.com/high-moctane/milbot/botplugins/restart"
	_ "github.com/joho/godotenv/autoload"
)

// plugins にプラグインを入れていくぞ(｀･ω･´)！
var plugins = []botplugin.Plugin{
	atnd.New(),
	exit.New(),
	kitakunoki.New(),
	ping.New(),
	restart.New(),
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ログ
	log.Print("milbot launch (｀･ω･´)！")
	botlog.Send("milbot launch (｀･ω･´)")
	defer log.Print("milbot terminated (｀･ω･´)")
	defer botlog.Send("milbot terminated (｀･ω･´)")

	// Bot の起動
	errCh := make(chan error)
	bot := NewBot(plugins)
	go func() { errCh <- bot.Serve(ctx) }()
	defer bot.Stop()

	// シグナルハンドリングや Bot のエラーによる終了処理
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-errCh:
		return err
	case sig := <-sigCh:
		switch sig {
		case syscall.SIGINT:
			botlog.Send("receive SIGINT")
		case syscall.SIGTERM:
			botlog.Send("received SITGERM")
		}
	}

	return nil
}
