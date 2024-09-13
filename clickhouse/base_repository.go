package clickhouse

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/bldsoft/gost/log"
)

type BaseRepository struct {
	db *Storage
}

func NewBaseRepository(storage *Storage) BaseRepository {
	return BaseRepository{db: storage}
}

func (r *BaseRepository) Storage() *Storage {
	return r.db
}

func (r *BaseRepository) RunSelect(ctx context.Context, query sq.SelectBuilder) (*sql.Rows, error) {
	r.LogQuery(ctx, query)
	return query.RunWith(r.Storage().Db).QueryContext(ctx)
}

func (r *BaseRepository) LogQuery(ctx context.Context, query sq.SelectBuilder) {
	str, params, _ := query.ToSql()
	log.FromContext(ctx).TracefWithFields(log.Fields{"params": params}, "Query: %s", str)
}
