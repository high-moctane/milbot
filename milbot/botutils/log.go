package botutils

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/nlopes/slack"
)

// logger は標準出力に吐くロガー
var logger = log.New(os.Stdout, "milbot: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)

// Log は標準出力にログを吐きます
func Log(v ...interface{}) { logger.Print(v...) }

// Logf はフォーマットを指定して標準出力にログを吐きます
func Logf(format string, v ...interface{}) { logger.Printf(format, v...) }

// LogWebhook は #milbot_log にログを送ります
func LogWebhook(v ...interface{}) {
	msg := fmt.Sprint(v...)
	if err := postWebhookLog(msg); err != nil {
		Log(v...)
	}
}

// LogWebhookf は #milbot_log にログを送るけどフォーマットを指定できます
func LogWebhookf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	if err := postWebhookLog(msg); err != nil {
		Logf(format, v...)
	}
}

// LogBoth は標準出力にも Webhook にもログを送信します
func LogBoth(v ...interface{}) {
	Log(v...)
	LogWebhook(v...)
}

// LogBothf はフォーマットを指定して標準出力にも Webhook にもログを送信します
func LogBothf(format string, v ...interface{}) {
	Logf(format, v...)
	LogWebhookf(format, v...)
}

// LogEventReceive は 誰からなんのメッセージを受け取ったかを標準出力に吐きます。
func LogEventReceive(api *slack.Client, event *slack.MessageEvent, description string) {
	username, err := GetUsername(api, event)
	if err != nil {
		username = "(could not get username)"
	}
	Log("received message about "+description+" by ", username)
}

// LogSendMessage はメッセージを送ったというログを残す
func LogSendMessage(channel, ts, text string) {
	Logf("sent message: {channel: %s, ts: %s, text: %s}", channel, ts, text)
}

// getURL は webhook の url を取得します
func getWebhookURL() (string, error) {
	url, ok := os.LookupEnv("SLACK_LOG_WEBHOOK_URL")
	if !ok {
		return "", fmt.Errorf("could not find SLACK_LOG_WEBHOOK_URL")
	}
	return url, nil
}

// postLog で msg を #milbot_log に投げます
func postWebhookLog(msg string) error {
	url, err := getWebhookURL()
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", makeWebhookBody(msg))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("postLog failed with status %s", resp.Status)
	}

	return nil
}

// makeWebhookBody は POST する JSON を生成します
func makeWebhookBody(msg string) *strings.Reader {
	return strings.NewReader(`{"text": "` + msg + `"}`)
}
