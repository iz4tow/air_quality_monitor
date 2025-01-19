package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"github.com/mdp/qrterminal"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
//	waProto "go.mau.fi/whatsmeow/binary/proto"
	waE2E "go.mau.fi/whatsmeow/proto/waE2E" // Updated import for waE2E
//	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"      // Ensure Protobuf is imported
	_ "github.com/mattn/go-sqlite3"         // SQLite driver
	waLog "go.mau.fi/whatsmeow/util/log"
)

func main() {
	msg:=os.Args[1]
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

	// Send a "Hello" message using waE2E.Message
	recipientJID := types.NewJID("393473055801", types.DefaultUserServer) //types.DefaultUserServer automatically adds @s.whatsapp.net to the JID.

	// Construct the message using waE2E.Message
	textMessage := &waE2E.Message{
		Conversation: proto.String(msg), // The actual message content
	}

	// Send the message with SendRequestExtra
	sendResp, err := client.SendMessage(context.Background(), recipientJID, textMessage) // Pass `nil` for optional extra params
	if err != nil {
		log.Fatalf("Failed to send message: %v", err)
	}
	fmt.Printf("Message sent successfully: %+v\n", sendResp)

	// Wait to ensure the message gets delivered
	time.Sleep(5 * time.Second)
}

