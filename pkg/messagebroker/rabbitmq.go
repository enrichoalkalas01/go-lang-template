package messagebroker

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type RabbitMQConfig struct {
	URL   string
	VHost string
}

type RabbitMQ struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
	log     *zap.Logger
}

func NewRabbitMQ(env *viper.Viper, log *zap.Logger) (*RabbitMQ, error) {
	config := RabbitMQConfig{
		URL:   env.GetString("RABBITMQ_URL"),
		VHost: env.GetString("RABBITMQ_VHOST"),
	}

	// Connect to RabbitMQ
	conn, err := amqp.Dial(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	// Open channel
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open rabbitmq channel: %w", err)
	}

	log.Info("RabbitMQ connected successfully",
		zap.String("url", config.URL),
	)

	return &RabbitMQ{
		Conn:    conn,
		Channel: ch,
		log:     log,
	}, nil
}

func (r *RabbitMQ) Close() error {
	if err := r.Channel.Close(); err != nil {
		return err
	}
	return r.Conn.Close()
}

// DeclareQueue creates a queue
func (r *RabbitMQ) DeclareQueue(name string) (amqp.Queue, error) {
	return r.Channel.QueueDeclare(
		name,  // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
}
