package botutils

import (
	"fmt"

	"github.com/nlopes/slack"
)

// GetUsername はユーザ名を取得します
func GetUsername(api *slack.Client, event *slack.MessageEvent) (string, error) {
	user, err := api.GetUserInfo(event.User)
	username := user.Name
	if err != nil {
		return "", fmt.Errorf("could not get username: %s", err)
	}
	return username, nil
}

// SendMessageWithLog はログ付きでメッセージを送ります
func SendMessageWithLog(api *slack.Client, event *slack.MessageEvent, message string) {
	channel, ts, text, err := api.SendMessage(
		event.Channel,
		slack.MsgOptionText(message, true),
	)
	if err != nil {
		LogBoth("post error: ", err)
		return
	}
	LogSendMessage(channel, ts, text)
}
