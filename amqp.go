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
	exchangeInfo struct {
		exname string
		extype string
	}
	// AMQPConnection is the main struct to manage a connection to an AMQP server.
	// It keeps exchanges and queues up and register AMQP publishers and subscribers.
	AMQPConnection struct {
		URI        string
		Connection *amqp.Connection
		Channel    *amqp.Channel
		// keeps information on initialized Exchanges
		// to be used to initialize Publishers: we need only name and type.
		exchanges map[string]exchangeInfo

		// keeps initialized amqp Queues
		// to be used to initialize Subscribers
		queues map[string]amqp.Queue

		// publishers to be send messages to the relative exchanges
		Publishers map[string]*AMQPPublisher

		// subscribers to start and listen for messages from relative queues
		Subscribers map[string]*AMQPSubscriber
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
	return &AMQPConnection{
		URI:         URI,
		exchanges:   make(map[string]exchangeInfo),
		queues:      make(map[string]amqp.Queue),
		Publishers:  make(map[string]*AMQPPublisher),
		Subscribers: make(map[string]*AMQPSubscriber),
	}
}

// Connect open an AMQP connection and setup the channel.
func (conn *AMQPConnection) Connect() error {
	var err error

	if conn.Connection, err = amqp.Dial(conn.URI); err != nil {
		return err
	}

	if conn.Channel, err = conn.Connection.Channel(); err != nil {
		return err
	}

	if err = conn.declareExchanges(); err != nil {
		return err
	}

	if err = conn.declareQueues(); err != nil {
		return err
	}

	if err = conn.initPublishers(); err != nil {
		return err
	}

	return nil
}

// declareExchanges will setup each of the configured Exchanges
func (conn *AMQPConnection) declareExchanges() error {
	for _, exchange := range amqpConf.Exchanges {
		err := conn.Channel.ExchangeDeclare(
			exchange.Name,
			exchange.Type,
			exchange.Durable,
			exchange.AutoDelete,
			exchange.Internal,
			exchange.NoWait,
			nil)

		if err != nil {
			return err
		}

		// add exchange info into the exchanges map
		ei := exchangeInfo{
			exname: exchange.Name,
			extype: exchange.Type,
		}
		conn.exchanges[ei.exname] = ei
	}
	return nil
}

func (conn *AMQPConnection) declareQueues() error {
	for _, queue := range amqpConf.Queues {
		q, err := conn.Channel.QueueDeclare(
			queue.Name,
			queue.Durable,
			queue.AutoDelete,
			queue.Exclusive,
			queue.NoWait,
			nil,
		)

		if err != nil {
			return err
		}

		// add the queue into the queues map
		conn.queues[q.Name] = q
	}
	return nil
}

func (conn *AMQPConnection) initPublishers() error {
	for _, p := range amqpConf.Publishers {
		if conn.Publishers[p.Name] == nil {
			publisher, err := conn.NewPublisher(p.Name)
			if err != nil {
				return err
			}
			conn.Publishers[p.Name] = publisher
		}

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

// NewPublisher return a new AMQP Publisher object initialized with "name"-publisher configuration.
func (conn *AMQPConnection) NewPublisher(name string) (*AMQPPublisher, error) {
	if conn.Publishers[name] != nil {
		return conn.Publishers[name], nil
	}

	// get publisher configuration
	p, ok := publisherFor(name)

	if !ok {
		return nil, fmt.Errorf("no configuration fond for publisher '%s'", name)
	}

	// initialize and register the AMQP Publisher struct
	ei := conn.exchanges[p.Exchange]
	publisher := &AMQPPublisher{
		Connection:   conn,
		Exchange:     ei.exname,
		ExchangeType: ei.extype,
		RoutingKey:   p.RoutingKey,
		ContentType:  p.ContentType,
		DeliveryMode: p.DeliveryMode,
	}

	// conn.publishers[p.Name] = publisher

	return publisher, nil

	// if p, ok := publisherFor(name); ok {
	// 	ei := conn.exchanges[p.Exchange]
	// 	publisher := &AMQPPublisher{
	// 		Connection:   conn,
	// 		Exchange:     ei.exname,
	// 		ExchangeType: ei.extype,
	// 		RoutingKey:   p.RoutingKey,
	// 		ContentType:  p.ContentType,
	// 		DeliveryMode: p.DeliveryMode,
	// 	}

	// 	conn.publishers[p.Name] = publisher

	// 	return publisher, nil

	// 	// if exchange, ok := exchangeFor(publisher.Name); ok {
	// 	// 	err := conn.Channel.ExchangeDeclare(
	// 	// 		exchange.Name,
	// 	// 		exchange.Type,
	// 	// 		exchange.Durable,
	// 	// 		exchange.AutoDelete,
	// 	// 		exchange.Internal,
	// 	// 		exchange.NoWait,
	// 	// 		nil)

	// 	// 	if err != nil {
	// 	// 		return nil, err
	// 	// 	}

	// 	// 	return &AMQPPublisher{
	// 	// 		Connection:   conn,
	// 	// 		Exchange:     exchange.Name,
	// 	// 		ExchangeType: exchange.Type,
	// 	// 		RoutingKey:   publisher.RoutingKey,
	// 	// 		ContentType:  publisher.ContentType,
	// 	// 		DeliveryMode: publisher.DeliveryMode,
	// 	// 	}, nil
	// 	// }
	// }

	// return nil, nil
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
