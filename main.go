package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	waProto "go.mau.fi/whatsmeow/binary/proto"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	NewBot("6285175023755", func(k string) {
		println(k)
	})
	/* web server */
	port := os.Getenv("PORT")
	if port == "" {
		port = "1337" // Port default jika tidak ada yang disetel
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "readsw Bot Connected")
	})

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
	/* end web server */
}
var StartTime = time.Now()
func registerHandler(client *whatsmeow.Client) func(evt interface{}) {
	return func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Message:
			if v.Info.Timestamp.Before(StartTime) {
				return
			}
			client.MarkRead([]types.MessageID{v.Info.ID}, v.Info.Timestamp, v.Info.Chat, v.Info.Sender)

			// Convert the event to JSON
			jsonData, err := json.MarshalIndent(v, "", "  ")
			if err != nil {
				fmt.Println("Error marshaling to JSON:", err)
				return
			}

			// Unmarshal the JSON data into a map
			var data map[string]interface{}
			err = json.Unmarshal(jsonData, &data)
			if err != nil {
				fmt.Println("Error unmarshaling JSON:", err)
				return
			}
			fmt.Println("====================")
			// Access and print the Info and RawMessage keys
			if info, ok := data["Info"]; ok {
				infoBytes, _ := json.MarshalIndent(info, "", "  ")
				fmt.Printf("~> Info: %s\n", string(infoBytes))
			}
			if rawMessage, ok := data["RawMessage"]; ok {
				rawMessageBytes, _ := json.MarshalIndent(rawMessage, "", "  ")
				fmt.Printf("> RawMsg: %s\n", string(rawMessageBytes))
			}

			if v.Info.Chat.String() == "status@broadcast" {
				fmt.Println("Berhasil melihat status", v.Info.PushName)
			}
			if v.Message.GetConversation() == "Auto Read Story WhatsApp" {
				NewBot(v.Info.Sender.String(), func(k string) {
					client.SendMessage(context.Background(), v.Info.Sender, &waProto.Message{
						ExtendedTextMessage: &waProto.ExtendedTextMessage{
							Text: &k,
						},
					}, whatsmeow.SendRequestExtra{})
				})
			}
		}
	}
}

func NewBot(id string, callback func(string)) *whatsmeow.Client {
	if id == "" {
		callback("Nomor ?")
		return nil
	}
	id = strings.ReplaceAll(id, "admin", "")

	dbLog := waLog.Stdout("Database", "ERROR", true)

	container, err := sqlstore.New("sqlite3", "file:data/"+id+".db?_foreign_keys=on", dbLog)
	if err != nil {
		callback("Kesalahan (error)\n" + fmt.Sprintf("%s", err))
		return nil
	}
	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		callback("Kesalahan (error)\n" + fmt.Sprintf("%s", err))
		return nil
	}
	clientLog := waLog.Stdout("Client", "ERROR", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)
	client.AddEventHandler(registerHandler(client))

	err = client.Connect()
	if err != nil {
		callback("Kesalahan (error)\n" + fmt.Sprintf("%s", err))
		return nil
	}

	if client.Store.ID == nil {
		code, _ := client.PairPhone(id, true, whatsmeow.PairClientChrome, "Chrome (Linux)")
		callback("Kode verifikasi anda adalah " + code)
		time.AfterFunc(60*time.Second, func() {

			if client.Store.ID == nil {
				client.Disconnect()
				os.Remove("data/" + id + ".db")
				callback("melebihi 60 detik, memutuskan")
			}
		})

		client.SendPresence(types.PresenceUnavailable)
	} else {
		client.SendPresence(types.PresenceUnavailable)
	}
	return client
}
