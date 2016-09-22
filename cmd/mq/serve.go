package main

import "github.com/urfave/cli"

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
	panic("not yet implemented")
}
