package transcodely

import (
	"context"
	"errors"
	"testing"
)

// fakePages returns a pageFetcher that walks a fixed list of pages and records
// the cursor it was called with on each invocation.
func fakePages(pages [][]int, cursors []string) (pageFetcher[int], *[]string) {
	calls := []string{}
	i := 0
	return func(_ context.Context, cursor string) ([]int, string, error) {
		calls = append(calls, cursor)
		if i >= len(pages) {
			return nil, "", errors.New("ran out of pages")
		}
		items := pages[i]
		next := cursors[i]
		i++
		return items, next, nil
	}, &calls
}

func TestIter_SinglePage(t *testing.T) {
	fetch, calls := fakePages([][]int{{1, 2, 3}}, []string{""})
	it := newIter(context.Background(), fetch)

	var got []int
	for it.Next() {
		got = append(got, it.Current())
	}
	if err := it.Err(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := []int{1, 2, 3}; !equalInts(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
	if len(*calls) != 1 || (*calls)[0] != "" {
		t.Errorf("expected 1 call with empty cursor, got %v", *calls)
	}
}

func TestIter_MultiplePagesWithCursor(t *testing.T) {
	fetch, calls := fakePages(
		[][]int{{1, 2}, {3, 4}, {5}},
		[]string{"c1", "c2", ""},
	)
	it := newIter(context.Background(), fetch)

	var got []int
	for it.Next() {
		got = append(got, it.Current())
	}
	if err := it.Err(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := []int{1, 2, 3, 4, 5}; !equalInts(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
	if want := []string{"", "c1", "c2"}; !equalStrings(*calls, want) {
		t.Errorf("cursors = %v, want %v", *calls, want)
	}
}

func TestIter_EmptyFirstPageStopsImmediately(t *testing.T) {
	fetch, _ := fakePages([][]int{{}}, []string{""})
	it := newIter(context.Background(), fetch)
	if it.Next() {
		t.Errorf("Next() = true on empty page, want false")
	}
	if err := it.Err(); err != nil {
		t.Errorf("Err() = %v, want nil", err)
	}
}

func TestIter_FetcherErrorPropagates(t *testing.T) {
	want := errors.New("network error")
	fetch := pageFetcher[int](func(_ context.Context, _ string) ([]int, string, error) {
		return nil, "", want
	})
	it := newIter(context.Background(), fetch)
	if it.Next() {
		t.Errorf("Next() = true after fetcher error, want false")
	}
	if got := it.Err(); !errors.Is(got, want) {
		t.Errorf("Err() = %v, want %v", got, want)
	}
}

func TestIter_ErrorMidStreamHaltsIteration(t *testing.T) {
	want := errors.New("mid-stream failure")
	calls := 0
	fetch := pageFetcher[int](func(_ context.Context, _ string) ([]int, string, error) {
		calls++
		if calls == 1 {
			return []int{1, 2}, "c1", nil
		}
		return nil, "", want
	})
	it := newIter(context.Background(), fetch)
	var got []int
	for it.Next() {
		got = append(got, it.Current())
	}
	if want := []int{1, 2}; !equalInts(got, want) {
		t.Errorf("got %v before failure, want %v", got, want)
	}
	if !errors.Is(it.Err(), want) {
		t.Errorf("Err() = %v, want %v", it.Err(), want)
	}
}

func TestIter_CloseIsIdempotentNoop(t *testing.T) {
	fetch, _ := fakePages([][]int{{1}}, []string{""})
	it := newIter(context.Background(), fetch)
	if err := it.Close(); err != nil {
		t.Errorf("Close() = %v, want nil", err)
	}
	if err := it.Close(); err != nil {
		t.Errorf("Close() (2nd call) = %v, want nil", err)
	}
}

func TestIter_ContextPassedToFetcher(t *testing.T) {
	type ctxKey struct{}
	ctx := context.WithValue(context.Background(), ctxKey{}, "tag")
	saw := false
	fetch := pageFetcher[int](func(c context.Context, _ string) ([]int, string, error) {
		if c.Value(ctxKey{}) == "tag" {
			saw = true
		}
		return []int{1}, "", nil
	})
	it := newIter(ctx, fetch)
	for it.Next() {
		_ = it.Current()
	}
	if !saw {
		t.Errorf("fetcher did not receive the iterator's context")
	}
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
