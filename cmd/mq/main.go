package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/drone/mq/stomp"

	"github.com/urfave/cli"
)

var client *stomp.Client

func main() {
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "server, s",
			Usage:  "stomp server address",
			Value:  "tcp://localhost:9000",
			EnvVar: "STOMP_SERVER",
		},
		cli.StringFlag{
			Name:   "usernane, u",
			Usage:  "stomp server username",
			EnvVar: "STOMP_USERNAME",
		},
		cli.StringFlag{
			Name:   "password, p",
			Usage:  "stomp server password",
			EnvVar: "STOMP_PASSWORD",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:   "publish",
			Usage:  "publish to a topic",
			Action: send,
			Before: setup,
			After:  teardown,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "data, d",
					Usage: "load message body from a file",
				},
				cli.Int64Flag{
					Name:  "expires",
					Usage: "sends the message with an expiration",
				},
				cli.DurationFlag{
					Name:  "ttl",
					Usage: "sends the message with a ttl",
				},
				cli.StringSliceFlag{
					Name:  "H, header",
					Usage: "sends the message with a custom header",
				},
				cli.BoolFlag{
					Name:  "persist",
					Usage: "sends the message with persistence",
				},
				cli.StringFlag{
					Name:  "retain",
					Usage: "sends the message with retention",
				},
				cli.BoolFlag{
					Name:  "receipt",
					Usage: "sends the message with request request",
				},
			},
		},
		{
			Name:   "subscribe",
			Usage:  "subscribe to a topic",
			Action: subscribe,
			Before: setup,
			After:  teardown,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "where",
					Usage: "subscribes to topics matching the SQL filter",
				},
				cli.IntFlag{
					Name:  "prefetch",
					Usage: "subscribes with a prefetch limit",
				},
				cli.StringFlag{
					Name:  "ack",
					Usage: "subscribes with ack settings",
				},
			},
		},
		comandServe,
		comandBench,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

// setup creates the STOMP client
func setup(c *cli.Context) (err error) {
	client, err = createClient(c)
	return
}

// create and connect the STOMP client.
func createClient(c *cli.Context) (*stomp.Client, error) {
	target := c.GlobalString("server")

	cli, err := stomp.Dial(target)
	if err != nil {
		return nil, err
	}

	if err := cli.Connect(); err != nil {
		return nil, err
	}
	return cli, nil
}

// teardown disconnects the STOMP client
func teardown(c *cli.Context) (err error) {
	if client != nil {
		err = client.Disconnect()
	}
	return
}

// send publishes a message to the specified topic.
func send(c *cli.Context) (err error) {
	var (
		path = c.Args().First()
		args = c.Args().Get(1)
	)

	var opts []stomp.MessageOption
	if expires := c.Int64("expires"); expires != 0 {
		opts = append(opts, stomp.WithExpires(expires))
	}
	if ttl := c.Duration("ttl"); ttl != 0 {
		exp := time.Now().Add(ttl).Unix()
		opts = append(opts, stomp.WithExpires(exp))
	}
	if c.Bool("receipt") {
		opts = append(opts, stomp.WithReceipt())
	}
	if c.Bool("persist") {
		opts = append(opts, stomp.WithPersistence())
	}
	if retain := c.String("retain"); retain != "" {
		opts = append(opts, stomp.WithRetain(retain))
	}
	if headers := c.StringSlice("H"); len(headers) != 0 {
		for _, header := range headers {
			parts := strings.SplitN(header, ":", 2)
			if len(parts) == 2 {
				opts = append(opts, stomp.WithHeader(parts[0], parts[1]))
			}
		}
	}

	return client.Send(path, []byte(args), opts...)
}

// subscribe subscribes to the specified topic.
func subscribe(c *cli.Context) (err error) {
	var (
		path = c.Args().First()
		quit = make(chan os.Signal, 1)
	)

	var opts []stomp.MessageOption
	if prefetch := c.Int("prefetch"); prefetch != 0 {
		opts = append(opts, stomp.WithPrefetch(prefetch))
	}
	if where := c.String("where"); where != "" {
		opts = append(opts, stomp.WithSelector(where))
	}
	if ack := c.String("ack"); ack != "" {
		opts = append(opts, stomp.WithAck(ack))
	}

	handler := func(m *stomp.Message) {
		log.Println(m)
		m.Release()
	}

	id, err := client.Subscribe(path, stomp.HandlerFunc(handler), opts...)
	if err != nil {
		return err
	}

	// block and listen for events until we get ctrl+c
	// forcing a friendly goodbye.
	signal.Notify(quit, os.Interrupt)

	<-quit
	return client.Unsubscribe(id)
}
