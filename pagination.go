package transcodely

import "context"

// Iter is the iterator returned by every List* method. It transparently fetches
// further pages on demand so callers can write a single for-loop:
//
//	iter := client.Jobs.List(ctx, &transcodely.JobListParams{Limit: 50})
//	for iter.Next() {
//	    job := iter.Current()
//	    // ...
//	}
//	if err := iter.Err(); err != nil {
//	    // ...
//	}
//
// To bound iteration, break out of the loop or call Close.
type Iter[T any] struct {
	ctx     context.Context
	fetch   pageFetcher[T]
	current T
	buffer  []T
	cursor  string
	hasMore bool
	primed  bool
	err     error
}

// pageFetcher returns one page of items plus the cursor for the next page.
// nextCursor is the empty string when no further pages exist.
type pageFetcher[T any] func(ctx context.Context, cursor string) (items []T, nextCursor string, err error)

func newIter[T any](ctx context.Context, fetch pageFetcher[T]) *Iter[T] {
	return &Iter[T]{ctx: ctx, fetch: fetch, hasMore: true}
}

// Next loads the next item, fetching another page if necessary. It returns
// false when there are no more items or when an error occurs (check Err()).
func (it *Iter[T]) Next() bool {
	if it.err != nil {
		return false
	}
	if len(it.buffer) > 0 {
		it.current = it.buffer[0]
		it.buffer = it.buffer[1:]
		return true
	}
	if !it.hasMore && it.primed {
		return false
	}
	items, next, err := it.fetch(it.ctx, it.cursor)
	it.primed = true
	if err != nil {
		it.err = err
		return false
	}
	it.cursor = next
	it.hasMore = next != ""
	if len(items) == 0 {
		return false
	}
	it.current = items[0]
	it.buffer = items[1:]
	return true
}

// Current returns the most recent item produced by Next.
// Calling Current before Next has returned true is undefined.
func (it *Iter[T]) Current() T { return it.current }

// Err returns the first error encountered during iteration, or nil.
func (it *Iter[T]) Err() error { return it.err }

// Close releases iterator state. The current implementation is a no-op (HTTP
// connections close automatically), but you should still defer Close so the
// signature is stable across future versions.
func (it *Iter[T]) Close() error { return nil }
