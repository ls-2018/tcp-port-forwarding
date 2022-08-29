package lib

import (
	"fmt"
	"net/url"

	"github.com/streadway/amqp"
)

var connMap map[string]*amqp.Connection = make(map[string]*amqp.Connection)

var channelMap map[string]*amqp.Channel = make(map[string]*amqp.Channel)

var Subscribed []string = make([]string, 0)

// 创建通道
func Channel(URL, QName string) (*amqp.Channel, error) {
	if IsNil(connMap[URL]) {
		p, e := url.Parse(URL)
		if e != nil {
			fmt.Printf("[ERR] AMQPURL: %s , %s", URL, e)
			return nil, e
		}
		// p.Path
		con, e := amqp.DialConfig(URL, amqp.Config{Vhost: p.Path[1:]})
		if e != nil {
			fmt.Printf("[ERR] AMQP: %s , %s", URL, e)
			return nil, e
		}
		connMap[URL] = con
	}
	channel := URL + QName
	if IsNil((channelMap[channel])) {
		ch, e := connMap[URL].Channel()
		if e != nil {
			return nil, e
		}
		channelMap[channel] = ch
	}
	return channelMap[channel], nil
}

// 发布内容
func Publish(URL, QName, Data string) error {
	c, e := Channel(URL, QName)
	if e == nil {
		return c.Publish("", QName, false, false, amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(Data),
		})
	}
	return e
}

// 监听回调
func Subscribe(URL, QName, Name string, cb func(string)) error {
	Key := URL + QName
	for _, v := range Subscribed {
		if v == Key {
			return nil
		}
	}
	c, e := Channel(URL, QName)
	if e == nil {
		ch, e := c.Consume(QName, Name, true, false, false, false, amqp.Table{})
		if e != nil {
			return e
		}
		Subscribed = append(Subscribed, Key)
		// return chan
		go (func() {
			for {
				c := <-ch
				cb(string(c.Body))
			}
		})()
	}
	return e
}
