package main

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"os"
	"path"

	"github.com/tidwall/redlog"
	"github.com/urfave/cli"
	"golang.org/x/crypto/acme/autocert"

	"github.com/drone/mq/logger"
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
			Name:   "cert",
			Usage:  "stomp ssl cert",
			EnvVar: "STOMP_CERT",
		},
		cli.StringFlag{
			Name:   "key",
			Usage:  "stomp ssl key",
			EnvVar: "STOMP_KEY",
		},
		cli.BoolFlag{
			Name:   "lets-encrypt",
			Usage:  "stomp ssl using lets encrypt",
			EnvVar: "STOMP_LETS_ENCRYPT",
		},
		cli.StringFlag{
			Name:   "lets-encrypt-host",
			Usage:  "stomp lets encrypt host",
			EnvVar: "STOMP_LETS_ENCRYPT_HOST",
		},
		cli.StringFlag{
			Name:   "lets-encrypt-email",
			Usage:  "stomp lets encrypt email",
			EnvVar: "STOMP_LETS_ENCRYPT_EMAIL",
		},
		cli.StringFlag{
			Name:   "lets-encrypt-cache",
			Usage:  "stomp lets encrypt cache directory",
			EnvVar: "STOMP_LETS_ENCRYPT_DIR",
		},
		cli.StringFlag{
			Name:   "base, b",
			Usage:  "stomp server base",
			Value:  "/",
			EnvVar: "STOMP_BASE",
		},
		cli.IntFlag{
			Name:   "level",
			Usage:  "logging level",
			Value:  2,
			EnvVar: "STOMP_LOG_LEVEL",
		},
	},
}

func serve(c *cli.Context) error {
	var (
		errc = make(chan error)

		user  = c.GlobalString("username")
		pass  = c.GlobalString("password")
		addr1 = c.String("tcp")
		addr2 = c.String("http")
		base  = c.String("base")
		route = c.String("path")
		cert  = c.String("cert")
		key   = c.String("key")

		acme  = c.Bool("lets-encrypt")
		host  = c.String("lets-encrypt-host")
		email = c.String("lets-encrypt-email")
		cache = c.String("lets-encrypt-cache")
	)

	var opts []server.Option
	if user != "" || pass != "" {
		opts = append(opts,
			server.WithCredentials(user, pass),
		)
	}

	logs := redlog.New(os.Stderr)
	logs.SetLevel(
		c.Int("level"),
	)
	logger.SetLogger(logs)
	logger.Noticef("stomp: starting server")

	server := server.NewServer(opts...)
	http.HandleFunc(path.Join("/", base, "meta/sessions"), server.HandleSessions)
	http.HandleFunc(path.Join("/", base, "meta/destinations"), server.HandleDests)
	http.Handle(path.Join("/", base, route), server)

	go func() {
		switch {
		case acme:
			errc <- listendAndServeAcme(host, email, cache)
		case cert != "":
			errc <- http.ListenAndServeTLS(addr2, cert, key, nil)
		default:
			errc <- http.ListenAndServe(addr2, nil)
		}
	}()

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

// helper function to setup and http server using let's encrypt
// certificates with auto-renewal.
func listendAndServeAcme(host, email, cache string) error {
	m := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(host),
		Email:      email,
	}
	if cache != "" {
		m.Cache = autocert.DirCache(cache)
	}
	s := &http.Server{
		Addr:      ":https",
		TLSConfig: &tls.Config{GetCertificate: m.GetCertificate},
	}
	return s.ListenAndServeTLS("", "")
}
