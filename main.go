package main

import (
	"fmt"
	"sync"

	"github.com/gofiber/fiber/v2"
)

type Message struct {
	Data string `json:"data"`
}

type PubSub struct {
	subs []chan Message
	mu   sync.Mutex
}

// Subscribe adds a new subscriber channel
func (ps *PubSub) Subscribe() chan Message {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ch := make(chan Message, 1)
	ps.subs = append(ps.subs, ch)
	return ch
}

// Publish sends a message to all subscriber channels
func (ps *PubSub) Publish(msg *Message) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	for _, sub := range ps.subs {
		sub <- *msg
	}
}

// Unsubscribe removes a subscriber channel
func (ps *PubSub) Unsubscribe(ch chan Message) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	for index, sub := range ps.subs {
		if sub == ch {
			ps.subs = append(ps.subs[:index], ps.subs[index+1:]...)
			close(ch)
			break
		}
	}
}

func main() {
	app := fiber.New()

	pubsub := &PubSub{}

	app.Post("/publisher", func(c *fiber.Ctx) error {
		msg := new(Message)
		if err := c.BodyParser(msg); err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		pubsub.Publish(msg)
		return c.JSON(&fiber.Map{
			"message": "Add to subscriber",
		})
	})

	sub := pubsub.Subscribe()
	go func() {
		for msg := range sub {
			fmt.Println("Receive message: ", msg)
		}
	}()

	sub2 := pubsub.Subscribe()
	go func() {
		for msg := range sub2 {
			fmt.Println("Receive message: ", msg)
		}
	}()
	// pubsub.Unsubscribe(pubsub.subs[1])

	app.Listen("localhost:8080")
}

func bakcUp() {
	channel1 := make(chan int)
	channel2 := make(chan int)

	go func() {
		channel1 <- 10
		close(channel1)
	}()

	go func() {
		channel2 <- 20
		close(channel2)
	}()

	closedChannel1, closedChannel2 := false, false
	for {
		if closedChannel1 && closedChannel2 {
			break // Stop loop
		}

		select {
		case value, ok := <-channel1:
			if !ok {
				closedChannel1 = true
				continue // Skip current loop
			}
			fmt.Println("Channel1: ", value)
		case value, ok := <-channel2:
			if !ok {
				closedChannel1 = true
				continue
			}
			fmt.Println("Channel2: ", value)
		}
	}
}
