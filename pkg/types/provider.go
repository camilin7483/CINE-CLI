package types

import "context"

type Provider interface {
	Name() string
	Priority() int
	Search(ctx context.Context, query string) ([]MediaRef, error)
	GetStream(ctx context.Context, ref MediaRef) (*Stream, error)
}
