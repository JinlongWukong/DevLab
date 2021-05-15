package notification

import (
	"encoding/json"
	"log"
	"os"

	"github.com/JinlongWukong/CloudLab/utils"
)

//Message channel buffer
var MessageChan = make(chan MessageRequest, 1000)

// MessageCreateRequest is the Create Message Request Parameters
type MessageRequest struct {
	RoomID        string   `json:"roomId,omitempty"`        // Room ID.
	ToPersonID    string   `json:"toPersonId,omitempty"`    // Person ID (for type=direct).
	ToPersonEmail string   `json:"toPersonEmail,omitempty"` // Person email (for type=direct).
	Text          string   `json:"text,omitempty"`          // Message in plain text format.
	Markdown      string   `json:"markdown,omitempty"`      // Message in markdown format.
	Files         []string `json:"files,omitempty"`         // File URL array.
}

//notficaiont internal api interface
func SendNotification(message MessageRequest) {

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
			payload, _ := json.Marshal(message)
			err, _ := utils.HttpSendJsonDataWithAuthBearer("https://webexapis.com/v1/messages", "POST", myToken, payload)
			if err != nil {
				log.Println(err)
				continue
			}
		}
	}()

	return

}
