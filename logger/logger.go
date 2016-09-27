package logger

var std Logger = new(none)

// Debugf writes a debug message to the standard logger.
func Debugf(format string, args ...interface{}) {
	std.Debugf(format, args...)
}

// Verbosef writes a verbose message to the standard logger.
func Verbosef(format string, args ...interface{}) {
	std.Verbosef(format, args...)
}

// Noticef writes a notice message to the standard logger.
func Noticef(format string, args ...interface{}) {
	std.Noticef(format, args...)
}

// Warningf writes a warning message to the standard logger.
func Warningf(format string, args ...interface{}) {
	std.Warningf(format, args...)
}

// Printf writes a default message to the standard logger.
func Printf(format string, args ...interface{}) {
	std.Printf(format, args...)
}

// SetLogger sets the standard logger.
func SetLogger(logger Logger) {
	std = logger
}

// Logger represents a logger.
type Logger interface {

	// Debugf writes a debug message.
	Debugf(string, ...interface{})

	// Verbosef writes a verbose message.
	Verbosef(string, ...interface{})

	// Noticef writes a notice message.
	Noticef(string, ...interface{})

	// Warningf writes a warning message.
	Warningf(string, ...interface{})

	// Printf writes a default message.
	Printf(string, ...interface{})
}

// none is a logger that silently ignores all writes.
type none struct{}

func (*none) Debugf(string, ...interface{})   {}
func (*none) Verbosef(string, ...interface{}) {}
func (*none) Noticef(string, ...interface{})  {}
func (*none) Warningf(string, ...interface{}) {}
func (*none) Printf(string, ...interface{})   {}
