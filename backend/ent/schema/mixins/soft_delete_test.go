package mixins_test

import (
	"context"
	"testing"

	"entgo.io/ent"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"
)

func TestSoftDeleteInterceptorAcceptsGeneratedQuery(t *testing.T) {
	interceptors := (mixins.SoftDeleteMixin{}).Interceptors()
	if len(interceptors) != 1 {
		t.Fatalf("interceptor count = %d, want 1", len(interceptors))
	}

	traverser, ok := interceptors[0].(ent.Traverser)
	if !ok {
		t.Fatalf("interceptor %T does not implement ent.Traverser", interceptors[0])
	}
	query := dbent.NewClient().User.Query()
	if err := traverser.Traverse(context.Background(), query); err != nil {
		t.Fatalf("traverse generated query: %v", err)
	}
}
