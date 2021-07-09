package notification

import (
	"context"
	"log"
	"os"
	"sync"

	"github.com/JinlongWukong/DevLab/config"
	"github.com/JinlongWukong/DevLab/manager"
)

var notificationKind = "webex"
var queueSize = 1000
var myToken string
var notifer notification = WebexNotification{}

// Message channel buffer
var messageChan = make(chan Message, queueSize)

// Message format
type Message struct {
	Target string `json:"target"` // Target ID.
	Text   string `json:"text"`   // Message in plain text format.
}

type Notifier struct {
}

var _ manager.Manager = Notifier{}

type notification interface {
	sendMessage(Message) error
}

func init() {
	// notification kind
	if config.Notification.Kind != "" {
		notificationKind = config.Notification.Kind
	}
	if notificationKind == "webex" {
		notifer = WebexNotification{}
	} else if notificationKind == "telegram" {
		notifer = telegramNotification{}
	} else {
		log.Printf("notification kind %v not supported", notificationKind)
	}
	// notification queue size
	if config.Notification.QueueSize > 0 {
		messageChan = make(chan Message, queueSize)
	}
	// bot token
	myToken = os.Getenv("BOT_TOKEN")
}

//notification internal api interface
func SendNotification(message Message) {

	// if token not define, notifer do not work, return
	if myToken == "" {
		return
	}

	select {
	case messageChan <- message:
	default:
		log.Println("Message buffer is full!!!")
	}

}

//controller loop
func (n Notifier) Control(ctx context.Context, wg *sync.WaitGroup) {
	log.Println("Notification manager started")
	defer func() {
		log.Println("Notification manager exited")
		wg.Done()
	}()

	if myToken == "" {
		log.Println("Error: bot token not found")
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case message := <-messageChan:
			if err := notifer.sendMessage(message); err != nil {
				log.Println(err.Error())
			}
		}
	}
}
