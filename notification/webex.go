package notification

import (
	"encoding/json"

	"github.com/JinlongWukong/DevLab/utils"
)

// WebexMessageRequest is the Webex Teams Create Message Request Parameters
type WebexNotification struct {
}

type WebexMessageRequest struct {
	RoomID        string   `json:"roomId,omitempty"`        // Room ID.
	ToPersonID    string   `json:"toPersonId,omitempty"`    // Person ID (for type=direct).
	ToPersonEmail string   `json:"toPersonEmail,omitempty"` // Person email (for type=direct).
	Text          string   `json:"text,omitempty"`          // Message in plain text format.
	Markdown      string   `json:"markdown,omitempty"`      // Message in markdown format.
	Files         []string `json:"files,omitempty"`         // File URL array.
}

func (webex WebexNotification) sendMessage(message Message) error {
	webexMessage := WebexMessageRequest{ToPersonEmail: message.Target, Text: message.Text}
	payload, _ := json.Marshal(webexMessage)
	if err, _ := utils.HttpSendJsonDataWithAuthBearer("https://webexapis.com/v1/messages", "POST", myToken, payload); err != nil {
		return err
	}
	return nil
}
