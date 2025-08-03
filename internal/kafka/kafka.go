package kafka

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
)

// Producer wraps a Sarama SyncProducer for sending messages
type Producer struct {
	producer sarama.SyncProducer
}

// NewProducer creates a new Kafka producer
func NewProducer(brokers []string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}
	return &Producer{producer: producer}, nil
}

// SendMessage sends a message to the specified topic
func (p *Producer) SendMessage(ctx context.Context, topic, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, _, err = p.producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(data),
	})
	return err
}

// Close closes the producer
func (p *Producer) Close() error {
	return p.producer.Close()
}

// Consumer wraps a Sarama ConsumerGroup for consuming messages
type Consumer struct {
	group sarama.ConsumerGroup
}

// NewConsumer creates a new Kafka consumer group
func NewConsumer(brokers []string, groupID string) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	group, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}
	return &Consumer{group: group}, nil
}

// Consume starts consuming messages from the specified topics
func (c *Consumer) Consume(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error {
	return c.group.Consume(ctx, topics, handler)
}

// Close closes the consumer group
func (c *Consumer) Close() error {
	return c.group.Close()
}
