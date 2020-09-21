package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type PublishRequest struct {
	Host string `json:"host"`
	Port uint16 `json:"port"`
	Username *string `json:"username"`
	Password *string `json:"password"`
	Topic *string `json:"topic"`
	Payload string `json:"payload"`
}

func defaultPublishRequest() PublishRequest {
	return PublishRequest{
		Host:     "localhost",
		Port:     1883,
	}
}

func handlePublish(w http.ResponseWriter, r *http.Request) {
	pr := defaultPublishRequest()
	err := json.NewDecoder(r.Body).Decode(&pr)
	if err != nil || pr.Topic == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	opts := mqtt.NewClientOptions().AddBroker(fmt.Sprintf("tcp://%s:%d", pr.Host, pr.Port)).SetClientID("mgott")
	opts.SetKeepAlive(2 * time.Second)

	if pr.Username != nil && pr.Password != nil {
		opts.Username = *pr.Username
		opts.Password = *pr.Password
	}

	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.WaitTimeout(time.Second * 5) && token.Error() != nil {
		http.Error(w, token.Error().Error(), http.StatusBadGateway)
		log.Println(token.Error())
		return
	}

	if token := c.Publish(*pr.Topic, 2, false, pr.Payload); token.WaitTimeout(time.Second * 5) && token.Error() != nil {
		http.Error(w, token.Error().Error(), http.StatusBadGateway)
		log.Println(token.Error())
		return
	}

	c.Disconnect(0)
}

func main() {
	http.HandleFunc("/publish", handlePublish)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
