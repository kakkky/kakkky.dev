package errors

import (
	"context"
	"log/slog"
)

type holder struct{ err error }

type ctxKey struct{}

// NewContext は err を stash する holder を注入した派生 ctx を返す。ErrorHandler middleware で呼ぶ。
func NewContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKey{}, &holder{})
}

// Set は ctx 内の holder に err を set する。
func Set(ctx context.Context, err error) {
	h, ok := ctx.Value(ctxKey{}).(*holder)
	if !ok {
		slog.Error("errors.Set: ctx に holder が無い (ErrorHandler middleware で wrap されていない可能性)", "err", err)
		return
	}
	h.err = err
}

// FromContext は holder に stash された err を返す。無ければ nil。
func FromContext(ctx context.Context) error {
	h, ok := ctx.Value(ctxKey{}).(*holder)
	if !ok {
		return nil
	}
	return h.err
}
