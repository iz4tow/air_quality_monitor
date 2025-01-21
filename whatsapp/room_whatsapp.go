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
	"bufio"
	"net/http"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"strconv"
	"strings"
	"flag"
)

type SensorData struct {
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	CO2         float64 `json:"co2"`
	NH3         float64 `json:"nh3"`
	NOx         float64 `json:"nox"`
	Dust25      float64 `json:"pm2.5"`
	Dust10      float64 `json:"pm10"`
	CO          float64 `json:"CO"`
}

var WhatsmeowClient *whatsmeow.Client
var wa_contact,password string

func main() {
	flag.StringVar(&wa_contact, "number","", "Whatsapp contact number whitout +, es 393312345654")
	flag.StringVar(&password,"password", "", "A secret word that allow any contact to receive sensor data")
	flag.Parse()
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
	container, err := sqlstore.New("sqlite3", "file:accounts2.db?_foreign_keys=on", dbLog)
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

//READ SENSOR DATA
func CreateReply(host string)(string){
	if (host == ""){
		// Read sensor host in pipefile if not provided
		pipe, err := os.Open("/tmp/airmonpipe")
		if err != nil {
			return "No pipe file!"
		}
		defer pipe.Close()
		reader := bufio.NewReader(pipe)
		host, _ = reader.ReadString('\n')
	}
	fmt.Println(host)
	var failure int = 0
	var response string
	var data SensorData
	for failure<20{
		resp, err := http.Get(fmt.Sprintf("http://%s/api/data", host))
		if err != nil {
			response="Sensor connection error!"
			failure++
			continue
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			response="Malformed Json"
			failure++
			continue
		}
		if err := json.Unmarshal(body, &data); err != nil {
			response="Malformed Json"
			failure++
			continue
		}else{
			break
		}
	}
	response="Temperature: "+strconv.FormatFloat(data.Temperature,'f',2,64)+"\nHumidity: "+strconv.FormatFloat(data.Humidity,'f',2,64)+"\nPM2.5: "+strconv.FormatFloat(data.Dust25,'f',2,64)+"\nPM10: "+strconv.FormatFloat(data.Dust10,'f',2,64)+"\nCO2: "+strconv.FormatFloat(data.CO2,'f',2,64)+"\nNH3: "+strconv.FormatFloat(data.NH3,'f',2,64)+"\nNOx: "+strconv.FormatFloat(data.NOx,'f',2,64)+"\nCO: "+strconv.FormatFloat(data.CO,'f',2,64)
	return response
}


func HandleMessage(messageEvent *events.Message) {
	recipientJID := types.NewJID(wa_contact, types.DefaultUserServer) //types.DefaultUserServer automatically adds @s.whatsapp.net to the JID. es 393334455666
////////fmt.Printf("Message structure: %+v\n", messageEvent.Message)
	var messageContent string
	if messageEvent.Message.Conversation != nil { //old whatsapp version
		messageContent = messageEvent.Message.GetConversation()
	} else if messageEvent.Message.ExtendedTextMessage != nil { //new whatsapp version
		messageContent = messageEvent.Message.ExtendedTextMessage.GetText()
	}
	if ((messageContent == "status" || messageContent == "Status") && messageEvent.Info.Chat==recipientJID){
		log.Println("Status request received")
		reply := CreateReply("")
		WhatsmeowClient.SendMessage(context.Background(), recipientJID, &waE2E.Message{
			Conversation: &reply,
		})
	}else if(strings.HasPrefix(strings.ToLower(messageContent), strings.ToLower("host")) && messageEvent.Info.Chat==recipientJID){
		messageContent = messageContent[len("host "):]
		reply := CreateReply(messageContent)
		WhatsmeowClient.SendMessage(context.Background(), recipientJID, &waE2E.Message{
			Conversation: &reply,
		})
	}else if((messageContent == "help" || messageContent == "Help") && messageEvent.Info.Chat==recipientJID){
		reply:="To get sensor data simply write: status\nTo get sensor data from specific host: host <ip>"
		WhatsmeowClient.SendMessage(context.Background(), recipientJID, &waE2E.Message{
			Conversation: &reply,
		})
	}else if(messageContent == password && password != ""){
		reply := CreateReply("")
		WhatsmeowClient.SendMessage(context.Background(), messageEvent.Info.Chat, &waE2E.Message{
			Conversation: &reply,
		})
	}
}

