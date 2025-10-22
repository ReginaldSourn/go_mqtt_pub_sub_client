package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	// CLI flags (non-TLS by default)
	broker := flag.String("broker", "tcp://localhost:1883", "MQTT broker URL (tcp://host:port)")
	topic := flag.String("topic", "test/topic", "MQTT topic")
	clientID := flag.String("client-id", "go-mqtt-client", "MQTT client id")
	mode := flag.String("mode", "pubsub", "mode: pub | sub | pubsub")
	msg := flag.String("msg", "hello from go mqtt", "message to publish")
	qos := flag.Int("qos", 1, "MQTT QoS")
	flag.Parse()

	opts := mqtt.NewClientOptions()
	opts.AddBroker(*broker)f
	opts.SetClientID(*clientID)
	// MQTT 5 support: set protocol version to 5
	opts.SetProtocolVersion(5)

	// default handlers
	opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
		log.Printf("[default handler] topic=%s payload=%s qos=%d retained=%v\n", msg.Topic(), string(msg.Payload()), msg.Qos(), msg.Retained())
	})

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("connect error: %v", token.Error())
	}
	log.Printf("connected to %s", *broker)

	// graceful shutdown
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)

	// subscribe
	if *mode == "sub" || *mode == "pubsub" {
		subToken := client.Subscribe(*topic, byte(*qos), func(client mqtt.Client, msg mqtt.Message) {
			log.Printf("received topic=%s payload=%s qos=%d retained=%v\n", msg.Topic(), string(msg.Payload()), msg.Qos(), msg.Retained())
		})
		subToken.Wait()
		if subToken.Error() != nil {
			log.Fatalf("subscribe error: %v", subToken.Error())
		}
		log.Printf("subscribed to %s", *topic)
	}

	// publish
	if *mode == "pub" || *mode == "pubsub" {
		token := client.Publish(*topic, byte(*qos), false, *msg)
		token.Wait()
		if token.Error() != nil {
			log.Fatalf("publish error: %v", token.Error())
		}
		log.Printf("published message to %s: %s", *topic, *msg)
	}

	// wait until interrupt
	<-sigc
	log.Println("disconnecting...")
	client.Disconnect(250)
}
