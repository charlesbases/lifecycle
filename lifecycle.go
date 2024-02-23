package lifecycle

import (
	"context"
	"strings"
	"sync/atomic"

	"github.com/pkg/errors"
)

const (
	stopped = iota
	starting
	started
	stopping
)

// A Hook is a pair of start and stop callbacks, either of which can be nil.
type Hook struct {
	OnStart func(ctx context.Context) error
	OnStop  func(ctx context.Context) error
}

// Lifecycle .
type Lifecycle interface {
	Append(h ...Hook)
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// New .
func New() Lifecycle {
	return new(lifecycle)
}

var _ Lifecycle = (*lifecycle)(nil)

// lifecycle .
type lifecycle struct {
	state int32
	hooks []Hook
}

// Append adds a Hook to the lifecycle.
func (l *lifecycle) Append(h ...Hook) {
	l.hooks = append(l.hooks, h...)
}

// Start runs all OnStart hooks, returning immediately if it encounters an error
func (l *lifecycle) Start(ctx context.Context) error {
	if ctx == nil {
		return errors.New("start lifecycle with nil context")
	}
	if !atomic.CompareAndSwapInt32(&l.state, stopped, starting) {
		return errors.Errorf("attempted to start lifecycle when in state: %d", l.state)
	}

	for _, hook := range l.hooks {
		if hook.OnStart == nil {
			continue
		}
		if err := l.runStartHook(ctx, hook); err != nil {
			return err
		}
	}

	atomic.CompareAndSwapInt32(&l.state, starting, started)
	return nil
}

// runStartHook .
func (l *lifecycle) runStartHook(ctx context.Context, hook Hook) error {
	return hook.OnStart(ctx)
}

// Stop runs any OnStop hooks whose OnStart counterpart succeeded. OnStop
// hooks run in reverse order.
func (l *lifecycle) Stop(ctx context.Context) error {
	if ctx == nil {
		return errors.New("stop lifecycle with nil context")
	}
	if !atomic.CompareAndSwapInt32(&l.state, started, stopping) {
		return nil
	}

	var errs multierr = make([]error, 0, len(l.hooks))
	for _, hook := range l.hooks {
		if hook.OnStop == nil {
			continue
		}
		if err := l.runStopHook(ctx, hook); err != nil {
			errs = append(errs, err)
		}
	}

	atomic.CompareAndSwapInt32(&l.state, stopping, stopped)
	return errs
}

// runStopHook .
func (l *lifecycle) runStopHook(ctx context.Context, hook Hook) error {
	return hook.OnStop(ctx)
}

type multierr []error

// Error .
func (e multierr) Error() string {
	var buff strings.Builder
	for i := range e {
		if i != 0 {
			buff.WriteString("; ")
		}
		buff.WriteString(e[i].Error())
	}
	return buff.String()
}
