package bibleit

import (
	"context"
	"errors"
)

// SSHClient executes commands through a caller-owned SSH connection.
type SSHClient struct {
	execute func(context.Context, Command) (CommandResponse, error)
}

// NewSSHClient binds command helpers to an explicit caller-owned SSH executor.
// Use NewClient for resource operations over HTTP.
func NewSSHClient(execute func(context.Context, Command) (CommandResponse, error)) (*SSHClient, error) {
	if execute == nil {
		return nil, errors.New("an SSH command executor is required")
	}
	return &SSHClient{execute: execute}, nil
}

func (c *SSHClient) Execute(ctx context.Context, command Command) (CommandResponse, error) {
	if c.execute == nil {
		return CommandResponse{}, errors.New("an SSH command executor is required")
	}
	if command.wire == "" {
		return CommandResponse{}, errors.New("invalid zero-value command")
	}
	if command.streaming {
		return CommandResponse{}, errors.New("streaming commands require a persistent SSH subscription")
	}
	if len(command.wire) > MaxCommandBytes {
		return CommandResponse{}, &RequestSizeError{Part: "command", Bytes: len(command.wire), Limit: MaxCommandBytes}
	}
	if err := ctx.Err(); err != nil {
		return CommandResponse{}, err
	}
	return c.execute(ctx, command)
}

func (c *SSHClient) Read(ctx context.Context, ref Reference) (CommandResponse, error) {
	command, err := ReadCommand(ref)
	if err != nil {
		return CommandResponse{}, err
	}
	return c.Execute(ctx, command)
}
func (c *SSHClient) Search(ctx context.Context, translation, query string) (CommandResponse, error) {
	command, err := SearchCommand(translation, query)
	if err != nil {
		return CommandResponse{}, err
	}
	return c.Execute(ctx, command)
}
func (c *SSHClient) Ping(ctx context.Context) (CommandResponse, error) {
	return c.Execute(ctx, PingCommand())
}
func (c *SSHClient) Identity(ctx context.Context) (CommandResponse, error) {
	return c.Execute(ctx, IdentityCommand())
}
func (c *SSHClient) ListTranslations(ctx context.Context) (CommandResponse, error) {
	return c.Execute(ctx, TranslationListCommand())
}
