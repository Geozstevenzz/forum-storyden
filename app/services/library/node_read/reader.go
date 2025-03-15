package node_read

import (
	"context"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/opt"
	"go.uber.org/zap"

	"github.com/Southclaws/storyden/app/resources/library"
	"github.com/Southclaws/storyden/app/resources/library/node_querier"
	"github.com/Southclaws/storyden/app/resources/pagination"
	"github.com/Southclaws/storyden/app/services/authentication/session"
)

type HydratedQuerier struct {
	logger     *zap.Logger
	session    *session.Provider
	nodereader *node_querier.Querier
}

func New(
	logger *zap.Logger,
	session *session.Provider,
	nodereader *node_querier.Querier,
) *HydratedQuerier {
	return &HydratedQuerier{
		logger:     logger,
		session:    session,
		nodereader: nodereader,
	}
}

func (q *HydratedQuerier) GetBySlug(ctx context.Context, qk library.QueryKey, sortChildrenBy opt.Optional[node_querier.ChildSortRule]) (*library.Node, error) {
	session := q.session.AccountMaybe(ctx)

	opts := []node_querier.Option{}

	if s, ok := session.Get(); ok {
		opts = append(opts, node_querier.WithVisibilityRulesApplied(&s.ID))
	} else {
		opts = append(opts, node_querier.WithVisibilityRulesApplied(nil))
	}

	sortChildrenBy.Call(func(v node_querier.ChildSortRule) {
		opts = append(opts, node_querier.WithSortChildrenBy(v))
	})

	n, err := q.nodereader.Get(ctx, qk, opts...)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return n, nil
}

func (q *HydratedQuerier) ListChildren(ctx context.Context, qk library.QueryKey, pp pagination.Parameters, sortChildrenBy opt.Optional[node_querier.ChildSortRule]) (*pagination.Result[*library.Node], error) {
	session := q.session.AccountMaybe(ctx)

	opts := []node_querier.Option{}

	if s, ok := session.Get(); ok {
		opts = append(opts, node_querier.WithVisibilityRulesApplied(&s.ID))
	} else {
		opts = append(opts, node_querier.WithVisibilityRulesApplied(nil))
	}

	sortChildrenBy.Call(func(v node_querier.ChildSortRule) {
		opts = append(opts, node_querier.WithSortChildrenBy(v))
	})

	r, err := q.nodereader.ListChildren(ctx, qk, pp, opts...)
	if err != nil {
		return nil, fault.Wrap(err, fctx.With(ctx))
	}

	return r, nil
}
