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
	}

	// AMQPSubscriber define the AMQP Subscriber structure.
	// TODO: complete the definition!!!
	AMQPSubscriber struct {
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
func (conn *AMQPConnection) Close() {
	conn.Channel.Close()
	conn.Connection.Close()
}

// NewAMQPPublisher return a new AMQP Publisher object.
func NewAMQPPublisher(connection *AMQPConnection, exchange string, exchangeType string, routingKey string) (*AMQPPublisher, error) {
	if err := connection.Channel.ExchangeDeclare(
		exchange, exchangeType,
		true, false, false, false, nil); err != nil {
		return nil, err
	}

	return &AMQPPublisher{
		Connection:   connection,
		Exchange:     exchange,
		ExchangeType: exchangeType,
		RoutingKey:   routingKey,
	}, nil
}

// Publish send a message to the AMQP Exchange.
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
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         payload,
			Timestamp:    time.Now(),
		})

	return err
}
