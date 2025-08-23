package process

import (
	"context"
	"io"
	"time"
)

type Command struct {
	alias         string
	ctx           context.Context
	name          string
	args          []string
	dir           string
	env           []string
	stdin         io.Reader
	timeout       time.Duration
	idleTimeout   time.Duration
	quietly       bool
	tty           bool
	outputHandler func(typ, line string)
}

func NewCommand(ctx context.Context, name string, args ...string) *Command {
	return &Command{
		ctx:  ctx,
		name: name,
		args: args,
	}
}

func (c *Command) Alias(alias string) *Command {
	c.alias = alias
	return c
}

func (c *Command) Path(dir string) *Command {
	c.dir = dir
	return c
}

func (c *Command) Env(env map[string]string) *Command {
	for k, v := range env {
		c.env = append(c.env, k+"="+v)
	}
	return c
}

func (c *Command) Input(reader io.Reader) *Command {
	c.stdin = reader
	return c
}

func (c *Command) Timeout(duration time.Duration) *Command {
	c.timeout = duration
	return c
}

func (c *Command) IdleTimeout(duration time.Duration) *Command {
	c.idleTimeout = duration
	return c
}

func (c *Command) Quietly() *Command {
	c.quietly = true
	return c
}

func (c *Command) Tty() *Command {
	c.tty = true
	return c
}

func (c *Command) OnOutput(handler func(typ, line string)) *Command {
	c.outputHandler = handler
	return c
}
