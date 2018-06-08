package actions

import (
	"context"
	"fmt"

	"github.com/jtopjian/yak/lib/connections"

	"github.com/sirupsen/logrus"
)

// BaseFields represents fields which are
// standard to all resources.
type BaseFields struct {
	// Name is the name of the resource. The value
	// will differ from resource to resource.
	Name string `mapstructure:"name" required:"true"`

	// State represents the state of the resource.
	// It can either be "present", "absent", "latest",
	// or a version number
	State string `mapstructure:"state" default:"present"`

	// Sudo is if the command requires sudo to run.
	Sudo bool `mapstructure:"sudo"`

	// Timeout is a timeout for the command.
	Timeout int `mapstructure:"timeout"`

	// conn is an internal field to hold a connection.
	conn connections.Connection

	// All BaseField resources also have a ContextLogger.
	ContextLogger
}

// ContextLogger represents fields and methods to help with logging.
type ContextLogger struct {
	ctx context.Context
}

func (logger *ContextLogger) setLogger(ctx context.Context, resource, name, state string) {
	if log, ok := ctx.Value("log").(*logrus.Entry); ok {
		log = log.WithFields(logrus.Fields{
			"resource": resource,
			"name":     name,
			"state":    state,
		})
		logger.ctx = context.WithValue(ctx, "log", log)
	}
}

func (logger ContextLogger) logInfo(v ...interface{}) {
	log := logger.ctx.Value("log").(*logrus.Entry)
	if log != nil {
		if len(v) == 1 {
			log.Info(fmt.Sprintf(v[0].(string)))
		} else {
			log.Info(fmt.Sprintf(v[0].(string), v[1:]...))
		}
	}
}

func (logger ContextLogger) logDebug(v ...interface{}) {
	log := logger.ctx.Value("log").(*logrus.Entry)
	if log != nil {
		if len(v) == 1 {
			log.Debug(fmt.Sprintf(v[0].(string)))
		} else {
			log.Debug(fmt.Sprintf(v[0].(string), v[1:]...))
		}
	}
}

func (logger ContextLogger) logError(v ...interface{}) {
	log := logger.ctx.Value("log").(*logrus.Entry)
	if log != nil {
		if len(v) == 1 {
			log.Error(fmt.Sprintf(v[0].(string)))
		} else {
			log.Error(fmt.Sprintf(v[0].(string), v[1:]...))
		}
	}
}
