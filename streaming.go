package transcodely

import (
	"context"
	"errors"
	"io"
	"time"
)

// Stream is the iterator returned by Watch RPCs. It auto-reconnects on
// transient network failures up to (max retries) times, transparently filters
// server heartbeats, and surfaces a single typed error via Err().
//
//	stream := client.Jobs.Watch(ctx, jobID)
//	defer stream.Close()
//	for stream.Next() {
//	    event := stream.Current()
//	    // ...
//	}
//	if err := stream.Err(); err != nil {
//	    // ...
//	}
type Stream[T any] struct {
	ctx       context.Context
	cancel    context.CancelFunc
	open      streamOpener[T]
	conn      streamConn[T]
	current   T
	err       error
	closed    bool
	retries   int
	maxRetry  int
	heartbeat func(T) bool
}

type streamConn[T any] interface {
	Recv() (T, error)
	Close() error
}

type streamOpener[T any] func(ctx context.Context) (streamConn[T], error)

func newStream[T any](ctx context.Context, maxRetry int, isHeartbeat func(T) bool, open streamOpener[T]) *Stream[T] {
	ctx, cancel := context.WithCancel(ctx)
	return &Stream[T]{
		ctx:       ctx,
		cancel:    cancel,
		open:      open,
		maxRetry:  maxRetry,
		heartbeat: isHeartbeat,
	}
}

// Next reads the next non-heartbeat event. Returns false when the stream
// terminates cleanly, or when an unrecoverable error occurs (check Err()).
func (s *Stream[T]) Next() bool {
	if s.closed {
		return false
	}
	for {
		if s.conn == nil {
			conn, err := s.open(s.ctx)
			if err != nil {
				if s.tryReconnect(err) {
					continue
				}
				s.err = err
				return false
			}
			s.conn = conn
		}
		msg, err := s.conn.Recv()
		if err != nil {
			s.closeConn()
			if errors.Is(err, io.EOF) {
				return false
			}
			if s.tryReconnect(err) {
				continue
			}
			s.err = err
			return false
		}
		if s.heartbeat != nil && s.heartbeat(msg) {
			continue
		}
		s.current = msg
		return true
	}
}

// Current returns the most recently received event.
func (s *Stream[T]) Current() T { return s.current }

// Err returns the first non-EOF error the stream encountered, or nil.
func (s *Stream[T]) Err() error { return s.err }

// Close cancels the underlying request and releases resources. Safe to call
// multiple times.
func (s *Stream[T]) Close() error {
	if s.closed {
		return nil
	}
	s.closed = true
	s.cancel()
	return s.closeConn()
}

func (s *Stream[T]) closeConn() error {
	if s.conn == nil {
		return nil
	}
	err := s.conn.Close()
	s.conn = nil
	return err
}

func (s *Stream[T]) tryReconnect(err error) bool {
	// Errors from the typed-error mapper come through as our SDK errors. Only
	// network/transient classes are worth reconnecting on.
	var apiConn *APIConnectionError
	if !errors.As(err, &apiConn) {
		return false
	}
	if s.retries >= s.maxRetry {
		return false
	}
	s.retries++
	// Backoff with light jitter — reconnect storms hurt the server.
	delay := time.Duration(s.retries) * 500 * time.Millisecond
	if delay > 4*time.Second {
		delay = 4 * time.Second
	}
	select {
	case <-s.ctx.Done():
		return false
	case <-time.After(delay):
		return true
	}
}
