package main

import (
	"io"
	"net"

	"github.com/urfave/cli"

	"github.com/drone/mq/server"
)

var comandServe = cli.Command{
	Name:   "start",
	Usage:  "start the message broker daemon",
	Action: serve,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:   "tcp",
			Usage:  "stomp tcp server address",
			Value:  ":9000",
			EnvVar: "STOMP_TCP",
		},
		cli.StringFlag{
			Name:   "http",
			Usage:  "stomp http server address",
			Value:  ":8000",
			EnvVar: "STOMP_HTTP",
		},
		cli.StringFlag{
			Name:   "base, b",
			Usage:  "stomp server base",
			Value:  "/",
			EnvVar: "STOMP_BASE",
		},
		cli.StringFlag{
			Name:   "path, p",
			Usage:  "stomp server path",
			Value:  "/ws",
			EnvVar: "STOMP_PATH",
		},
	},
}

func serve(c *cli.Context) error {
	var (
		errc = make(chan error)

		addr1 = c.String("tcp")
		// addr2 = c.String("http")
		// base  = c.String("base")
		// route = c.String("path")
	)

	server := server.NewServer()
	// http.Handle(path.Join("/", base, route), server)

	// go func() {
	// 	errc <- http.ListenAndServe(addr2, nil)
	// }()

	go func() {
		l, err := net.Listen("tcp", addr1)
		if err != nil {
			errc <- err
			return
		}
		defer l.Close()

		for {
			conn, err := l.Accept()
			if err == io.EOF {
				errc <- nil
				return
			}
			if err != nil {
				errc <- err
				return
			}
			go server.Serve(conn)
		}
	}()

	return <-errc
}
