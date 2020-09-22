package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type PublishRequest struct {
	Host     string  `json:"host"`
	Port     uint16  `json:"port"`
	Username *string `json:"username"`
	Password *string `json:"password"`
	Topic    *string `json:"topic"`
	Payload  string  `json:"payload"`
}

func (pr *PublishRequest) Endpoint() string {
	return fmt.Sprintf("%s:%d", pr.Host, pr.Port)
}

func defaultPublishRequest() PublishRequest {
	return PublishRequest{
		Host: "localhost",
		Port: 1883,
	}
}

type ConnectionMap struct {
	Connections map[string]mqtt.Client
	lock        sync.RWMutex
}

func (cm *ConnectionMap) Publish(pr PublishRequest) error {
	cm.lock.RLock()
	c, exist := cm.Connections[pr.Endpoint()]
	if exist {
		err := cm.Publish(pr)
		cm.lock.RUnlock()
		if err != nil {
			return err
		}
	} else {
		cm.lock.RUnlock()
		cm.lock.Lock()
		defer cm.lock.Unlock()

		opts := mqtt.NewClientOptions()
		opts.AddBroker(fmt.Sprintf("tcp://%s:%d", pr.Host, pr.Port))
		opts.SetClientID(fmt.Sprintf("mgott-%d", time.Now().Unix()))
		opts.SetKeepAlive(2 * time.Second)
		opts.SetAutoReconnect(false)
		opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
			cm.Delete(pr.Endpoint())
		})

		if pr.Username != nil && pr.Password != nil {
			opts.Username = *pr.Username
			opts.Password = *pr.Password
		}

		c := mqtt.NewClient(opts)
		if token := c.Connect(); token.WaitTimeout(5*time.Second) && token.Error() != nil {
			// http.Error(w, token.Error().Error(), http.StatusBadGateway)
			log.Println(token.Error())
			return token.Error()
		}

		cm[pr.Endpoint()] = Connection{
			client:      c,
			endpoint:    pr.Endpoint(),
			disconnectC: nil,
			lock:        sync.RWMutex{},
		}
	}

	return nil
}

func (cm *ConnectionMap) Delete(pr string) {
	cm.lock.Lock()
	delete(cm.Connections, pr)
	cm.lock.Unlock()
}

type Connection struct {
	client      mqtt.Client
	endpoint    string
	disconnectC *time.Timer
	lock        sync.RWMutex
}

func (c *Connection) Publish(pr PublishRequest) error {
	c.lock.Lock()
	c.disconnectC = time.AfterFunc(5*time.Minute, func() {

	})
	c.lock.Unlock()

	c.lock.RLock()
	defer c.lock.RUnlock()
	if token := c.client.Publish(*pr.Topic, 2, false, pr.Payload); token.WaitTimeout(5*time.Second) && token.Error() != nil {
		log.Println(token.Error())
		return token.Error()
	}
}

func (c *Connection) Disconnect(pr PublishRequest) error {
	c.lock.Lock()
	c.client.Disconnect(0)
	c.lock.Unlock()

	connectionMap.Delete(c.endpoint)
}

var (
	connectionMap ConnectionMap
)

func handlePublish(w http.ResponseWriter, r *http.Request) {

	pr := defaultPublishRequest()
	err := json.NewDecoder(r.Body).Decode(&pr)
	if err != nil || pr.Topic == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", pr.Host, pr.Port))
	opts.SetClientID(fmt.Sprintf("mgott-%d", time.Now().Unix()))
	opts.SetKeepAlive(2 * time.Second)
	opts.SetAutoReconnect(false)

	if pr.Username != nil && pr.Password != nil {
		opts.Username = *pr.Username
		opts.Password = *pr.Password
	}

	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.WaitTimeout(5*time.Second) && token.Error() != nil {
		http.Error(w, token.Error().Error(), http.StatusBadGateway)
		log.Println(token.Error())
		return
	}

	if token := c.Publish(*pr.Topic, 2, false, pr.Payload); token.WaitTimeout(5*time.Second) && token.Error() != nil {
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
