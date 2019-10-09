// Copyright 2019 Luca Stasio <joshuagame@gmail.com>
// Copyright 2019 IT Resources s.r.l.
//
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package sgul defines common structures and functionalities for applications.
// amqp.go defines commons for amqp integration.
package sgul

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/streadway/amqp"
)

type (
	// AMQPConnection .
	AMQPConnection struct {
		URI        string
		Connection *amqp.Connection
		Channel    *amqp.Channel
	}
	// AMQPPublisher define the AMQP Publisher structure.
	// Normally can be used as a sort of repository by a business service.
	AMQPPublisher struct {
		Connection   *AMQPConnection
		Exchange     string
		ExchangeType string
		RoutingKey   string
		ContentType  string
		DeliveryMode uint8
	}

	// AMQPSubscriber define the AMQP Subscriber structure.
	// TODO: complete the definition!!!
	AMQPSubscriber struct {
		Connection *AMQPConnection
		Queue      string
		Consumer   string
		AutoAck    bool
		Exclusive  bool
		NoLocal    bool
		NoWait     bool
		Replies    <-chan amqp.Delivery
	}
)

var amqpConf = GetConfiguration().AMQP

// NewAMQPConnection return a new disconnected AMQP Connection structure.
func NewAMQPConnection() *AMQPConnection {
	URI := fmt.Sprintf("amqp://%s:%s@%s:%d/%s", amqpConf.User, amqpConf.Password, amqpConf.Host, amqpConf.Port, amqpConf.VHost)
	return &AMQPConnection{URI: URI}
}

// Connect open an AMQP connection and setup the channel.
func (conn *AMQPConnection) Connect() error {
	var err error

	conn.Connection, err = amqp.Dial(conn.URI)
	if err != nil {
		return err
	}

	conn.Channel, err = conn.Connection.Channel()
	if err != nil {
		return err
	}

	return nil
}

// Close closes AMQP channel and connection.
func (conn *AMQPConnection) Close() error {
	if err := conn.Channel.Close(); err != nil {
		return err
	}
	if err := conn.Connection.Close(); err != nil {
		return err
	}
	return nil
}

// // Publisher returns a new AMQP Publisher on this connection.
// func (conn *AMQPConnection) Publisher(exchange string, exchangeType string, routingKey string) (*AMQPPublisher, error) {
// 	return NewAMQPPublisher(conn, exchange, exchangeType, routingKey)
// }

// Subscriber reutrns a new AMQP Subscriber on this connection.
func (conn *AMQPConnection) Subscriber(queue string, consumer string, durable, autoDelete, autoAck, exclusive, noLocal, noWait bool) (*AMQPSubscriber, error) {
	return NewAMQPSubscriber(conn, queue, consumer, durable, autoDelete, autoAck, exclusive, noLocal, noWait)
}

// // NewAMQPPublisher return a new AMQP Publisher object.
// func NewAMQPPublisher(connection *AMQPConnection, exchange string, exchangeType string, routingKey string) (*AMQPPublisher, error) {
// 	if err := connection.Channel.ExchangeDeclare(
// 		exchange, exchangeType,
// 		true, false, false, false, nil); err != nil {
// 		return nil, err
// 	}

// 	return &AMQPPublisher{
// 		Connection:   connection,
// 		Exchange:     exchange,
// 		ExchangeType: exchangeType,
// 		RoutingKey:   routingKey,
// 	}, nil
// }

// NewPublisher return a new AMQP Publisher object initialized with "name"-publisher configuration.
func (conn *AMQPConnection) NewPublisher(name string) (*AMQPPublisher, error) {
	if publisher, ok := publisherFor(name); ok {
		if exchange, ok := exchangeFor(publisher.Name); ok {
			err := conn.Channel.ExchangeDeclare(
				exchange.Name,
				exchange.Type,
				exchange.Durable,
				exchange.AutoDelete,
				exchange.Internal,
				exchange.NoWait,
				nil)

			if err != nil {
				return nil, err
			}

			return &AMQPPublisher{
				Connection:   conn,
				Exchange:     exchange.Name,
				ExchangeType: exchange.Type,
				RoutingKey:   publisher.RoutingKey,
				ContentType:  publisher.ContentType,
				DeliveryMode: publisher.DeliveryMode,
			}, nil
		}
	}

	return nil, nil
}

func publisherFor(name string) (Publisher, bool) {
	for _, publisher := range amqpConf.Publishers {
		if strings.ToLower(publisher.Name) == strings.ToLower(name) {
			return publisher, true
		}
	}

	// no publisher configuration found for "name"
	return Publisher{}, false
}

func exchangeFor(name string) (Exchange, bool) {
	for _, exchange := range amqpConf.Exchanges {
		if exchange.Name == name {
			return exchange, true
		}
	}

	// no exchange configuration found for "name"
	return Exchange{}, false
}

// Publish send a message to the AMQP Exchange.
// TODO: try and parametrize even "mandatory" and "immdiate" flags.
func (pub *AMQPPublisher) Publish(event Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	err = pub.Connection.Channel.Publish(
		pub.Exchange,
		pub.RoutingKey,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: pub.DeliveryMode,
			ContentType:  pub.ContentType,
			Body:         payload,
			Timestamp:    time.Now(),
		})

	return err
}

// NewAMQPSubscriber returns a new AMQP Subscriber object.
func NewAMQPSubscriber(connection *AMQPConnection, queue string, consumer string, durable, autoDelete, autoAck, exclusive, noLocal, noWait bool) (*AMQPSubscriber, error) {
	q, err := connection.Channel.QueueDeclare(
		queue,
		durable,
		autoDelete,
		exclusive,
		noWait,
		nil,
	)

	if err != nil {
		return nil, err
	}

	return &AMQPSubscriber{
		Connection: connection,
		Queue:      q.Name,
		Consumer:   consumer,
		AutoAck:    autoAck,
		Exclusive:  exclusive,
		NoLocal:    noLocal,
		NoWait:     noWait,
	}, nil
}

// Consume start consuming messages from queue. Returns outputs channel to range on.
func (sub *AMQPSubscriber) Consume() (<-chan amqp.Delivery, error) {
	return sub.Connection.Channel.Consume(
		sub.Queue,
		sub.Consumer,
		sub.AutoAck,
		sub.Exclusive,
		sub.NoLocal,
		sub.NoWait,
		nil,
	)
}
