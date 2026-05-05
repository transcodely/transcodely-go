package transcodely

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"
)

// fakeConn is a streamConn that hands out a fixed sequence of (event, error)
// tuples. After the slice is exhausted it returns io.EOF.
type fakeConn[T any] struct {
	steps []fakeStep[T]
	idx   int
	closed bool
}

type fakeStep[T any] struct {
	event T
	err   error
}

func (f *fakeConn[T]) Recv() (T, error) {
	if f.idx >= len(f.steps) {
		var zero T
		return zero, io.EOF
	}
	s := f.steps[f.idx]
	f.idx++
	return s.event, s.err
}

func (f *fakeConn[T]) Close() error {
	f.closed = true
	return nil
}

func TestStream_YieldsEventsThenStopsOnEOF(t *testing.T) {
	conn := &fakeConn[int]{steps: []fakeStep[int]{{event: 1}, {event: 2}, {event: 3}}}
	open := func(_ context.Context) (streamConn[int], error) { return conn, nil }
	s := newStream(context.Background(), 0, nil, open)
	defer s.Close()

	var got []int
	for s.Next() {
		got = append(got, s.Current())
	}
	if err := s.Err(); err != nil {
		t.Fatalf("Err() = %v, want nil", err)
	}
	if want := []int{1, 2, 3}; !equalInts(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestStream_HeartbeatsAreFiltered(t *testing.T) {
	// Treat odd numbers as heartbeats; only even-numbered events should surface.
	conn := &fakeConn[int]{steps: []fakeStep[int]{
		{event: 2}, {event: 1}, {event: 4}, {event: 3}, {event: 6},
	}}
	open := func(_ context.Context) (streamConn[int], error) { return conn, nil }
	s := newStream(context.Background(), 0, func(n int) bool { return n%2 == 1 }, open)
	defer s.Close()

	var got []int
	for s.Next() {
		got = append(got, s.Current())
	}
	if want := []int{2, 4, 6}; !equalInts(got, want) {
		t.Errorf("got %v, want %v (heartbeats not filtered)", got, want)
	}
}

func TestStream_ReconnectsOnAPIConnectionError(t *testing.T) {
	// First connection drops with *APIConnectionError after one event; the
	// reconnect serves a snapshot and finishes cleanly.
	first := &fakeConn[int]{steps: []fakeStep[int]{
		{event: 1},
		{err: &APIConnectionError{baseError: baseError{msg: "dropped"}}},
	}}
	second := &fakeConn[int]{steps: []fakeStep[int]{{event: 99}}}
	calls := 0
	open := func(_ context.Context) (streamConn[int], error) {
		calls++
		if calls == 1 {
			return first, nil
		}
		return second, nil
	}
	// Cap reconnect delay tightly so the test stays fast (~500ms first retry).
	s := newStream(context.Background(), 1, nil, open)
	defer s.Close()

	var got []int
	for s.Next() {
		got = append(got, s.Current())
	}
	if err := s.Err(); err != nil {
		t.Fatalf("Err() = %v, want nil", err)
	}
	if want := []int{1, 99}; !equalInts(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
	if calls != 2 {
		t.Errorf("opener was called %d times, want 2 (initial + reconnect)", calls)
	}
}

func TestStream_DoesNotReconnectOnNonConnectionError(t *testing.T) {
	wantErr := &NotFoundError{baseError: baseError{msg: "missing"}}
	conn := &fakeConn[int]{steps: []fakeStep[int]{{err: wantErr}}}
	calls := 0
	open := func(_ context.Context) (streamConn[int], error) {
		calls++
		return conn, nil
	}
	s := newStream(context.Background(), 5, nil, open)
	defer s.Close()

	if s.Next() {
		t.Errorf("Next() = true on NotFoundError, want false")
	}
	var nf *NotFoundError
	if !errors.As(s.Err(), &nf) {
		t.Errorf("Err() = %T, want *NotFoundError", s.Err())
	}
	if calls != 1 {
		t.Errorf("opener was called %d times, want 1 (no reconnect)", calls)
	}
}

func TestStream_GivesUpAfterMaxReconnects(t *testing.T) {
	// Every connection drops immediately with a connection error.
	open := func(_ context.Context) (streamConn[int], error) {
		return &fakeConn[int]{steps: []fakeStep[int]{
			{err: &APIConnectionError{baseError: baseError{msg: "down"}}},
		}}, nil
	}
	// Use a context with deadline so a runaway never wedges the test.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s := newStream(ctx, 1, nil, open)
	defer s.Close()

	if s.Next() {
		t.Errorf("Next() = true after exceeding max reconnects, want false")
	}
	var conn *APIConnectionError
	if !errors.As(s.Err(), &conn) {
		t.Errorf("Err() = %T, want *APIConnectionError", s.Err())
	}
}

func TestStream_CloseIsIdempotent(t *testing.T) {
	open := func(_ context.Context) (streamConn[int], error) {
		return &fakeConn[int]{steps: []fakeStep[int]{{event: 1}}}, nil
	}
	s := newStream(context.Background(), 0, nil, open)

	if err := s.Close(); err != nil {
		t.Errorf("Close() = %v, want nil", err)
	}
	if err := s.Close(); err != nil {
		t.Errorf("Close() (2nd) = %v, want nil", err)
	}
	if s.Next() {
		t.Errorf("Next() after Close() should be false")
	}
}

func TestStream_OpenerErrorReconnectsWhenConnectionType(t *testing.T) {
	// Even an error from the *opener* should trigger a reconnect when typed
	// as *APIConnectionError.
	calls := 0
	open := func(_ context.Context) (streamConn[int], error) {
		calls++
		if calls == 1 {
			return nil, &APIConnectionError{baseError: baseError{msg: "no route"}}
		}
		return &fakeConn[int]{steps: []fakeStep[int]{{event: 7}}}, nil
	}
	s := newStream(context.Background(), 1, nil, open)
	defer s.Close()

	var got []int
	for s.Next() {
		got = append(got, s.Current())
	}
	if err := s.Err(); err != nil {
		t.Fatalf("Err() = %v, want nil", err)
	}
	if want := []int{7}; !equalInts(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
