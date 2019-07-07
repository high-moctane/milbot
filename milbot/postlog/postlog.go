package postlog

import (
	"fmt"
	"net/http"
	"os"
	"strings"
)

// Log は #milbot_log にログを送ります
func Log(v ...interface{}) error {
	msg := fmt.Sprint(v...)
	if err := postLog(msg); err != nil {
		return err
	}
	return nil
}

// Logf は #milbot_log にログを送るけどフォーマットを指定できます
func Logf(format string, a ...interface{}) error {
	msg := fmt.Sprintf(format, a...)
	if err := postLog(msg); err != nil {
		return err
	}
	return nil
}

// getURL は webhook の url を取得します
func getURL() (string, error) {
	url, ok := os.LookupEnv("SLACK_LOG_WEBHOOK_URL")
	if !ok {
		return "", fmt.Errorf("could not find SLACK_LOG_WEBHOOK_URL")
	}
	return url, nil
}

// postLog で msg を #milbot_log に投げます
func postLog(msg string) error {
	url, err := getURL()
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", makeBody(msg))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("postLog failed with status %s", resp.Status)
	}

	return nil
}

func makeBody(msg string) *strings.Reader {
	return strings.NewReader(`{"text": "` + msg + `"}`)
}
