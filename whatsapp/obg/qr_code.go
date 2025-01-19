package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"


	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	_ "github.com/mattn/go-sqlite3"
	waLog "go.mau.fi/whatsmeow/util/log"
	"github.com/mdp/qrterminal"
)


func CreateClient() (*whatsmeow.Client) {
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

var WhatsmeowClient *whatsmeow.Client
func main() {
	WhatsmeowClient = CreateClient()
	ConnectClient(WhatsmeowClient)

	WhatsmeowClient.Connect()

	// Listen to Ctrl+C (you can also do something else that prevents the program from exiting)
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	WhatsmeowClient.Disconnect()
}

