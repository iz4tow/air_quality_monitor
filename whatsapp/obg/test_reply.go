package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal"
	"go.mau.fi/whatsmeow"
	waE2E "go.mau.fi/whatsmeow/proto/waE2E" // Updated import for waE2E
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"go.mau.fi/whatsmeow/types"
)

var WhatsmeowClient *whatsmeow.Client

func main() {
	WhatsmeowClient = CreateClient()
	ConnectClient(WhatsmeowClient)
	WhatsmeowClient.AddEventHandler(HandleEvent)


	WhatsmeowClient.Connect()
	// Listen to Ctrl+C (you can also do something else that prevents the program from exiting)
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	WhatsmeowClient.Disconnect()
}

func CreateClient() *whatsmeow.Client {
	dbLog := waLog.Stdout("Database", "INFO", true)
	container, err := sqlstore.New("sqlite3", "file:accounts.db?_foreign_keys=on", dbLog)
	if err != nil {
		log.Fatalln(err)
	}

	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		log.Fatalln(err)
	}

	clientLog := waLog.Stdout("Client", "INFO", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	return client
}

func ConnectClient(client *whatsmeow.Client) {
	if client.Store.ID == nil {
		// No ID stored, new login, show a qr code
		qrChan, _ := client.GetQRChannel(context.Background())
		err := client.Connect()
		if err != nil {
			log.Fatalln(err)
		}

		for evt := range qrChan {
			if evt.Event == "code" {
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			} else {
				log.Println("Login event:", evt.Event)
			}
		}
	} else {
		// Already logged in, just connect
		err := client.Connect()
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func HandleEvent(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		go HandleMessage(v)
	}
}

func HandleMessage(messageEvent *events.Message) {
	wa_contact:=os.Args[1]
	recipientJID := types.NewJID(wa_contact, types.DefaultUserServer) //types.DefaultUserServer automatically adds @s.whatsapp.net to the JID. es 393334455666
	messageContent := messageEvent.Message.GetConversation()
	if (messageContent == "status" && messageEvent.Info.Chat==recipientJID){
		reply := "Hello World!"
		WhatsmeowClient.SendMessage(context.Background(), recipientJID, &waE2E.Message{
		//WhatsmeowClient.SendMessage(context.Background(), messageEvent.Info.Chat, &waE2E.Message{
			Conversation: &reply,
		})
	}
}

