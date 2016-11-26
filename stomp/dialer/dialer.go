package dialer

import (
	"net"
	"net/url"

	"golang.org/x/net/websocket"
)

const (
	protoHTTP  = "http"
	protoHTTPS = "https"
	protoWS    = "ws"
	protoWSS   = "wss"
	protoTCP   = "tcp"
)

// Dial creates a client connection to the given target.
func Dial(target string) (net.Conn, error) {
	u, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case protoHTTP, protoHTTPS, protoWS, protoWSS:
		return dialWebsocket(u)
	case protoTCP:
		return dialSocket(u)
	default:
		panic("stomp: invalid protocol")
	}
}

func dialWebsocket(target *url.URL) (net.Conn, error) {
	origin, err := target.Parse("/")
	if err != nil {
		return nil, err
	}
	switch origin.Scheme {
	case protoWS:
		origin.Scheme = protoHTTP
	case protoWSS:
		origin.Scheme = protoHTTPS
	}
	return websocket.Dial(target.String(), "", origin.String())
}

func dialSocket(target *url.URL) (net.Conn, error) {
	return net.Dial(protoTCP, target.Host)
}
