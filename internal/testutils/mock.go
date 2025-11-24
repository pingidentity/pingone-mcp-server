// Copyright © 2025 Ping Identity Corporation

package testutils

import (
	"context"
	"net/http"

	"github.com/patrickcping/pingone-go-sdk-v2/management"
	"github.com/pingidentity/pingone-go-client/pingone"
	"github.com/stretchr/testify/mock"
)

type MockPage[T any] struct {
	Data         *T
	HTTPResponse *http.Response
	Error        error
}

type LegacySdkMockPage struct {
	EntityArray  *management.EntityArray
	HTTPResponse *http.Response
	Error        error
}

var CancelledContextMatcher = mock.MatchedBy(func(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	// Check if the context is already cancelled
	select {
	case <-ctx.Done():
		return ctx.Err() == context.Canceled
	default:
		return false
	}
})

func MockPaginationIterator[T pingone.MappedNullable](pages []MockPage[T]) pingone.PagedIterator[T] {
	return func(yield func(pingone.PagedCursor[T], error) bool) {
		for _, page := range pages {
			cursor := pingone.PagedCursor[T]{
				Data:         page.Data,
				HTTPResponse: page.HTTPResponse,
			}

			if !yield(cursor, page.Error) {
				return
			}
		}
	}
}

func MockLegacySdkPaginationIterator(pages []LegacySdkMockPage) management.EntityArrayPagedIterator {
	return func(yield func(management.PagedCursor, error) bool) {
		for _, page := range pages {
			cursor := management.PagedCursor{
				EntityArray:  page.EntityArray,
				HTTPResponse: page.HTTPResponse,
			}

			if !yield(cursor, page.Error) {
				return
			}
		}
	}
}
