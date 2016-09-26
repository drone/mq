package stomp

import "golang.org/x/net/context"

const clientKey = "stomp.client"

// NewContext adds the client to the context.
func (c *Client) NewContext(ctx context.Context, client *Client) context.Context {
	// HACK for use with gin and echo
	if s, ok := ctx.(setter); ok {
		s.Set(clientKey, clientKey)
		return ctx
	}
	return context.WithValue(ctx, clientKey, client)
}

// FromContext retrieves the client from context
func FromContext(ctx context.Context) (*Client, bool) {
	client, ok := ctx.Value(clientKey).(*Client)
	return client, ok
}

// MustFromContext retrieves the client from context. Panics if not found
func MustFromContext(ctx context.Context) *Client {
	client, ok := FromContext(ctx)
	if !ok {
		panic("stomp.Client not found in context")
	}
	return client
}

// HACK setter defines a context that enables setting values. This is a
// temporary workaround for use with gin and echo and will eventually
// be removed. DO NOT depend on this.
type setter interface {
	Set(string, interface{})
}
