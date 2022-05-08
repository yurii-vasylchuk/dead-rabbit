package rabbitmq

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"

	"DeadRabbit/state"
)

type Configuration struct {
	User     string
	Password string
	Host     string
	Port     string
	Vhost    string
	Queue    string
	Dlq      string
}

func LoadMessages(c Configuration) ([]state.MessageStruct, error) {
	connectionString := fmt.Sprintf("amqp://%s:%s@%s:%s/%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Vhost)

	connection, err := amqp.Dial(connectionString)
	if err != nil {
		return nil, err
	}
	defer func(connection *amqp.Connection) {
		err := connection.Close()
		if err != nil {
			log.Printf("Can't close a connection, err is %v", err)
		}
	}(connection)

	channel, err := connection.Channel()
	if err != nil {
		return nil, err
	}
	defer func(channel *amqp.Channel) {
		err := channel.Close()
		if err != nil {
			log.Printf("Can't close a channel, err is %v", err)
		}
	}(channel)

	messages := make([]state.MessageStruct, 0, 10)

	for {
		if msg, ok, _ := channel.Get(c.Dlq, true); !ok {
			break
		} else {
			messages = append(messages, state.MessageStruct{
				Body:    string(msg.Body),
				Headers: msg.Headers,
			})
		}
	}

	return messages, nil
}

func PublishMessagesToDlq(messages []state.MessageStruct, c Configuration) error {
	connectionString := fmt.Sprintf("amqp://%s:%s@%s:%s/%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Vhost)

	connection, err := amqp.Dial(connectionString)
	if err != nil {
		return err
	}
	defer func(connection *amqp.Connection) {
		err := connection.Close()
		if err != nil {
			log.Printf("Can't close a connection, err is %v", err)
		}
	}(connection)

	channel, err := connection.Channel()
	if err != nil {
		return err
	}
	defer func(channel *amqp.Channel) {
		err := channel.Close()
		if err != nil {
			log.Printf("Can't close a channel, err is %v", err)
		}
	}(channel)

	for _, message := range messages {
		if err := channel.Publish("", c.Dlq, true, false, amqp.Publishing{
			Headers:     message.Headers,
			ContentType: "application/json",
			Body:        []byte(message.Body),
		}); err != nil {
			return err
		}
	}

	return nil
}

func PublishMessageToQueue(message state.MessageStruct, c Configuration) error {
	connectionString := fmt.Sprintf("amqp://%s:%s@%s:%s/%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Vhost)

	connection, err := amqp.Dial(connectionString)
	if err != nil {
		return err
	}
	defer func(connection *amqp.Connection) {
		err := connection.Close()
		if err != nil {
			log.Printf("Can't close a connection, err is %v", err)
		}
	}(connection)

	channel, err := connection.Channel()
	if err != nil {
		return err
	}
	defer func(channel *amqp.Channel) {
		err := channel.Close()
		if err != nil {
			log.Printf("Can't close a channel, err is %v", err)
		}
	}(channel)

	return channel.Publish("", c.Queue, true, false, amqp.Publishing{
		Headers:     message.Headers,
		ContentType: "application/json",
		Body:        []byte(message.Body),
	})
}
