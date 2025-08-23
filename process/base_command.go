package process

import (
	"io"
	"time"
)

type BaseCommand struct {
	alias         string
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

func (c *BaseCommand) Alias(alias string) *BaseCommand {
	c.alias = alias
	return c
}

func (c *BaseCommand) Path(dir string) *BaseCommand {
	c.dir = dir
	return c
}

func (c *BaseCommand) Env(env map[string]string) *BaseCommand {
	for k, v := range env {
		c.env = append(c.env, k+"="+v)
	}
	return c
}

func (c *BaseCommand) Input(reader io.Reader) *BaseCommand {
	c.stdin = reader
	return c
}

func (c *BaseCommand) Timeout(duration time.Duration) *BaseCommand {
	c.timeout = duration
	return c
}

func (c *BaseCommand) IdleTimeout(duration time.Duration) *BaseCommand {
	c.idleTimeout = duration
	return c
}

func (c *BaseCommand) Quietly() *BaseCommand {
	c.quietly = true
	return c
}

func (c *BaseCommand) Tty() *BaseCommand {
	c.tty = true
	return c
}

func (c *BaseCommand) OnOutput(handler func(typ, line string)) *BaseCommand {
	c.outputHandler = handler
	return c
}
