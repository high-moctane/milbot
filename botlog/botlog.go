package botlog

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/high-moctane/milbot/boterrors"
)

// envMilbotLogWebhookURL は #milbot_log に送信するための Webhook URL の
// 環境変数です。
const envMilbotLogWebhookURL = "MILBOT_LOG_WEBHOOK_URL"

// Send は stderr と #milbot_log にログを吐きます。
func Send(v ...interface{}) {
	log.Println(v...)
	if err := postMilbotLogWebhook(fmt.Sprint(v...)); err != nil {
		log.Println(err)
	}
}

// Sendf は stderr と #milbot_log にログを吐きます。Sprintf みたいな感じに
// 使います。
func Sendf(format string, v ...interface{}) {
	log.Printf(format, v...)
	if err := postMilbotLogWebhook(fmt.Sprintf(format, v...)); err != nil {
		log.Println(err)
	}
}

// postMilbotLogWebhook は msg を #milbot_log に送信します。
func postMilbotLogWebhook(msg string) error {
	url, err := getMilbotLogWebhookURL()
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", makeWebhookRequestBody(msg))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if _, err := io.Copy(ioutil.Discard, resp.Body); err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("postMilbotLogWebhook failed with status %s",
			resp.Status)
	}

	return nil
}

// makeWebhookRequestBody は Webhook に送信する POST リクエストの body を
// 作ります。
func makeWebhookRequestBody(msg string) *strings.Reader {
	return strings.NewReader(`{"text": "` + msg + `"}`)
}

// getWebhookURL は #milbot_log に送信できる Webhook の URL を環境変数から
// 取得します。取得できなかった場合は boterrors.NewErrInvalidEnv を返します。
func getMilbotLogWebhookURL() (url string, err error) {
	url, ok := os.LookupEnv(envMilbotLogWebhookURL)
	if !ok {
		err = boterrors.NewErrInvalidEnv(envMilbotLogWebhookURL)
		return
	}
	return
}
