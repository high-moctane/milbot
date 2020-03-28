package botlog

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

// envMilbotLogWebhookURL は #milbot_log に送信するための Webhook URL の
// 環境変数です。
const envMilbotLogWebhookURL = "MILBOT_LOG_WEBHOOK_URL"

// Send は #milbot_log にログを吐きます。
func Send(v ...interface{}) {
	SendContext(context.Background(), v...)
}

// SendContext は #milbot_log にログを吐きます。context.Context が
// 使えます。
func SendContext(ctx context.Context, v ...interface{}) {
	if err := postMilbotLogWebhookContext(ctx, fmt.Sprint(v...)); err != nil {
		log.Println(fmt.Errorf("can't send msg to #milbot_log: %w", err))
	}
}

// Sendf は #milbot_log にログを吐きます。Sprintf みたいな感じに
// 使います。
func Sendf(format string, v ...interface{}) {
	SendfContext(context.Background(), format, v...)
}

// SendfContext は #milbot_log にログを吐きます。Sprintf みたいな感じに
// 使います。context.Context が使えます。
func SendfContext(ctx context.Context, format string, v ...interface{}) {
	if err := postMilbotLogWebhookContext(ctx, fmt.Sprintf(format, v...)); err != nil {
		log.Println(fmt.Errorf("can't send msg to #milbot_log: %w", err))
	}
}

// postMilbotLogWebhook は msg を #milbot_log に送信します。
func postMilbotLogWebhookContext(ctx context.Context, msg string) error {
	url, err := getMilbotLogWebhookURL()
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, makeWebhookRequestBody(msg))
	if err != nil {
		return fmt.Errorf("can't create a request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("post error: %w", err)
	}
	defer resp.Body.Close()
	defer io.Copy(ioutil.Discard, resp.Body)

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
		err = errors.New(envMilbotLogWebhookURL + " not found")
		return
	}
	return
}
