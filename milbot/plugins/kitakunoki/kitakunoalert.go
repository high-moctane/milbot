package kitakunoki

import (
	"time"

	"github.com/high-moctane/milbot/milbot/botutils"
	"github.com/high-moctane/milbot/milbot/plugins/atnd"
	"github.com/nlopes/slack"
)

var kitakunoAlertHour = 21

func kitakunoAlertServer(api *slack.Client, plugin *Plugin) {
	for {
		<-time.NewTimer(kitakunoDuration()).C

		if atnd.Exist() {
			botutils.SendMessageWithLog(api, "#random", plugin.kitakunoMessage())
		}
	}
}

// kitakunoDuration は次の帰宅の木アラートまでの時間を取得します
func kitakunoDuration() time.Duration {
	now := time.Now()
	target := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		kitakunoAlertHour,
		0,
		0,
		0,
		time.Local,
	)
	dur := target.Sub(now)
	if dur < 0 {
		return dur + 24*time.Hour
	}
	return dur
}
