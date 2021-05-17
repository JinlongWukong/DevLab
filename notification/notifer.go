package notification

import (
	"encoding/json"
	"log"
	"os"

	"github.com/JinlongWukong/CloudLab/config"
	"github.com/JinlongWukong/CloudLab/utils"
)

var notificationKind = "webex"
var queueSize = 1000

//Message channel buffer
var MessageChan = make(chan Message, queueSize)

// MessageCreateRequest is the Create Message Request Parameters
type Message struct {
	Target string `json:"target,omitempty"` // Target ID.
	Text   string `json:"text,omitempty"`   // Message in plain text format.
}

//notficaiont internal api interface
func SendNotification(message Message) {

	select {
	case MessageChan <- message:
	default:
		log.Println("Message buffer is full!!!")
	}

}

//initialize configuration
func initialize() {
	if config.Notification.Kind != "" {
		notificationKind = config.Notification.Kind
	}
	if config.Notification.QueueSize > 0 {
		queueSize = config.Notification.QueueSize
	}
}

//controller loop
func Manager() {

	myToken := os.Getenv("BOT_TOKEN")

	go func() {
		for message := range MessageChan {
			formatedMessage := WebexMessageRequest{ToPersonEmail: message.Target, Text: message.Text}
			payload, _ := json.Marshal(formatedMessage)
			err, _ := utils.HttpSendJsonDataWithAuthBearer("https://webexapis.com/v1/messages", "POST", myToken, payload)
			if err != nil {
				log.Println(err)
			}
		}
	}()

	return

}
