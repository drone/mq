package server

// Option configures server options.
type Option func(*Server)

// WithAuth returns an Option which configures custom authorization
// for the STOMP server.
func WithAuth(auth Authorizer) Option {
	return func(s *Server) {
		s.router.authorizer = auth
	}
}

// WithCredentials returns an Option which configures basic authorization
// using the given username and password
func WithCredentials(username, password string) Option {
	return WithAuth(BasicAuth(username, password))
}
