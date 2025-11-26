package kafka

import (
	"log"

	"github.com/IBM/sarama"
)

type KafkaProducer struct {
	producer sarama.SyncProducer
	topic    string
}

func NewKafkaProducer(brokers []string, topic string) (*KafkaProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &KafkaProducer{
		producer: producer,
		topic:    topic,
	}, nil
}

func (p *KafkaProducer) Publish(message []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Value: sarama.ByteEncoder(message),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return err
	}

	log.Printf("🚀 Kafka: message sent | topic=%s partition=%d offset=%d\n", p.topic, partition, offset)
	return nil
}

func (p *KafkaProducer) Close() {
	p.producer.Close()
}
