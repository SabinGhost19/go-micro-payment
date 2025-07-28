package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/IBM/sarama"
)

type Order struct {
	CustomerName string `json:"customer_name"`
	ProductType  string `json:"product_type"`
}

func main() {
	http.HandleFunc("/", placeOrder)
	addr := ":3005"
	log.Fatal(http.ListenAndServe(addr, nil))
}

func connectToProducer(brokers []string) (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5

	return sarama.NewSyncProducer(brokers, config)
}
func sendToKafka(topic string, message []byte) error {
	brokers := []string{"localhost:9092"}
	producer, err := connectToProducer(brokers)
	if err != nil {
		return err
	}
	defer producer.Close()

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}
	partition, offser, err := producer.SendMessage(msg)
	if err != nil {
		return err
	}
	log.Printf("Message sent in topic %s to partition %d at offset %d\n", topic, partition, offser)

	return nil
}
func placeOrder(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	order := new(Order)
	if err := json.NewDecoder(r.Body).Decode(order); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	orderInBytes, err := json.Marshal(order)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := sendToKafka("product_demo_test", orderInBytes); err != nil {
		log.Printf("Failed to send message to Kafka: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"msg":     "Order placed successfully: " + order.CustomerName,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

}
