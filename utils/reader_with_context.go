package utils

import (
	"context"
	"io"
	"time"
)

type ReaderWithDeadline struct {
	r        io.Reader
	deadline time.Time
}

func NewReaderWithDeadline(r io.Reader, deadline time.Time) *ReaderWithDeadline {
	return &ReaderWithDeadline{
		r:        r,
		deadline: deadline,
	}
}

func (r *ReaderWithDeadline) Read(p []byte) (n int, err error) {
	if !r.deadline.IsZero() && r.deadline.Before(time.Now()) {
		return 0, context.DeadlineExceeded
	}
	return r.r.Read(p)
}

func ReadAllWithContext(ctx context.Context, r io.Reader) ([]byte, error) {
	deadline, _ := ctx.Deadline()
	r = NewReaderWithDeadline(r, deadline)
	return io.ReadAll(r)
}

type ReaderWithContext struct {
	r   io.Reader
	ctx context.Context
}

func NewReaderWithContext(ctx context.Context, r io.Reader) io.Reader {
	return &ReaderWithContext{r: r, ctx: ctx}
}

func (r *ReaderWithContext) Read(p []byte) (n int, err error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.r.Read(p)
}
