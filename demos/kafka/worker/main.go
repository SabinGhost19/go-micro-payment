package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/IBM/sarama"
)

func main() {

	topic := "product_demo_test"
	msgCount := 0

	worker, err := ConnectConsumer([]string{"localhost:9092"})
	if err != nil {
		panic(err)
	}

	consumer, err := worker.ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		panic(err)
	}
	log.Printf("Consumer started for topic %s", topic)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	doneChan := make(chan struct{})

	go func() {
		for {
			select {
			case err := <-consumer.Errors():
				fmt.Printf("Error: %v\n", err)
			case message := <-consumer.Messages():
				msgCount++
				log.Printf("Received message: %s, number: %d on topic: %s", string(message.Value), msgCount, topic)
				order := string(message.Value)
				log.Printf("Order received: %s...processing....", order)

			case <-signalChan:
				log.Println("Shutting down consumer...interrupt is detected")
				doneChan <- struct{}{}
			}
		}
	}()

	<-doneChan
	log.Printf("Total messages processed: %d", msgCount)
	if err := consumer.Close(); err != nil {
		panic(err)
	}

}

func ConnectConsumer(brokser []string) (sarama.Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	return sarama.NewConsumer(brokser, config)
}
