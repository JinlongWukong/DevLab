package notification

import (
	"github.com/JinlongWukong/DevLab/utils"
)

type telegramNotification struct {
}

type telegramMessageRequest struct {
	ChatID string `json:"chatId"` // Chat ID.
	Text   string `json:"text"`   // Message in plain text format.
}

func (telegram telegramNotification) sendMessage(message Message) error {
	telegramMessage := telegramMessageRequest{ChatID: message.Target, Text: message.Text}
	url := "https://api.telegram.org/bot" + myToken + "/sendMessage?" + "chat_id=" +
		telegramMessage.ChatID + "&text=" + telegramMessage.Text
	if err, _ := utils.HttpSendJsonData(url,
		"POST", nil); err != nil {
		return err
	}
	return nil
}
