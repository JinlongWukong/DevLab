package notification

import (
	"encoding/json"
	"log"
	"os"

	"github.com/JinlongWukong/CloudLab/utils"
)

//Message channel buffer
var MessageChan = make(chan Message, 1000)

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
