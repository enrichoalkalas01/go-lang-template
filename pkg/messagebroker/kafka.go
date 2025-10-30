package messagebroker

import (
	"fmt"
	"strings"

	"github.com/IBM/sarama"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type KafkaConfig struct {
	Brokers []string
	GroupID string
	Version string
}

type Kafka struct {
	Producer sarama.SyncProducer
	Consumer sarama.ConsumerGroup
	Client   sarama.Client
	log      *zap.Logger
}

func NewKafka(cfg *viper.Viper, log *zap.Logger) (*Kafka, error) {
	config := KafkaConfig{
		Brokers: strings.Split(cfg.GetString("KAFKA_BROKERS"), ","),
		GroupID: cfg.GetString("KAFKA_GROUP_ID"),
		Version: cfg.GetString("KAFKA_VERSION"),
	}

	// Parse Kafka version
	version, err := sarama.ParseKafkaVersion(config.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kafka version: %w", err)
	}

	// Configure Kafka
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = version
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	saramaConfig.Producer.Retry.Max = 5

	// Create client
	client, err := sarama.NewClient(config.Brokers, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka client: %w", err)
	}

	// Create producer
	producer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	// Create consumer group
	consumer, err := sarama.NewConsumerGroupFromClient(config.GroupID, client)
	if err != nil {
		producer.Close()
		client.Close()
		return nil, fmt.Errorf("failed to create kafka consumer: %w", err)
	}

	log.Info("Kafka connected successfully",
		zap.Strings("brokers", config.Brokers),
		zap.String("group_id", config.GroupID),
	)

	return &Kafka{
		Producer: producer,
		Consumer: consumer,
		Client:   client,
		log:      log,
	}, nil
}

func (k *Kafka) Close() error {
	if err := k.Producer.Close(); err != nil {
		return err
	}
	if err := k.Consumer.Close(); err != nil {
		return err
	}
	return k.Client.Close()
}
