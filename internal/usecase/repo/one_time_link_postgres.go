package repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/brshpl/otl/internal/entity"
	"github.com/brshpl/otl/pkg/postgres"
	"github.com/jackc/pgx/v4"
)

// OneTimeLinkPostgres stores entity.OneTimeLink-s in postgres
type OneTimeLinkPostgres struct {
	*postgres.Postgres
}

// NewPostgres - create new OneTimeLinkPostgres
func NewPostgres(pg *postgres.Postgres) *OneTimeLinkPostgres {
	return &OneTimeLinkPostgres{pg}
}

// Store entity.OneTimeLink in repo
func (r *OneTimeLinkPostgres) Store(ctx context.Context, t entity.OneTimeLink) error {
	sql, args, err := r.Builder.
		Insert("links").
		Columns("data, link, expired").
		Values(t.Data, t.Link, t.Expired).
		ToSql()
	if err != nil {
		return buildError("OneTimeLinkPostgres", "Store", "r.Builder", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return buildError("OneTimeLinkPostgres", "Store", "r.Pool.Exec", err)
	}

	return nil
}

// Get entity.OneTimeLink by link and expire this record. Returns empty entity.OneTimeLink if no data found
func (r *OneTimeLinkPostgres) Get(ctx context.Context, link string) (entity.OneTimeLink, error) {
	zero := entity.OneTimeLink{}

	sqlSelect, argsSelect, err := r.buildSelectSQL(link)
	if err != nil {
		return zero, buildError("OneTimeLinkPostgres", "Get", "r.Builder.Select", err)
	}

	sqlExpire, argsExpire, err := r.Builder.
		Update("links").
		Set("expired", true).
		Where(squirrel.Eq{"link": link}).
		ToSql()
	if err != nil {
		return zero, buildError("OneTimeLinkPostgres", "Get", "r.Builder.Update", err)
	}

	tx, err := r.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return zero, buildError("OneTimeLinkPostgres", "Get", "r.Pool.BeginTx", err)
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, sqlSelect, argsSelect...)

	otl := entity.OneTimeLink{}
	err = row.Scan(&otl.Data, &otl.Link, &otl.Expired)
	if errors.Is(err, pgx.ErrNoRows) {
		return zero, nil
	} else if err != nil {
		return zero, buildError("OneTimeLinkPostgres", "Get", "row.Scan", err)
	}

	if !otl.Expired {
		_, err = tx.Exec(ctx, sqlExpire, argsExpire...)
		if err != nil {
			return zero, buildError("OneTimeLinkPostgres", "Get", "r.Pool.Exec", err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return zero, buildError("OneTimeLinkPostgres", "Get", "tx.Commit", err)
	}

	return otl, nil
}

func (r *OneTimeLinkPostgres) Check(ctx context.Context, link string) (bool, error) {
	sql, args, err := r.buildSelectSQL(link)
	if err != nil {
		return false, buildError("OneTimeLinkPostgres", "Check", "r.Builder.Select", err)
	}

	row := r.Pool.QueryRow(ctx, sql, args...)

	var data string
	err = row.Scan(&data)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	} else if err != nil {
		return false, buildError("OneTimeLinkPostgres", "Check", "row.Scan", err)
	}

	return true, nil
}

func (r *OneTimeLinkPostgres) buildSelectSQL(link string) (string, []any, error) {
	return r.Builder.
		Select("data, link, expired").
		From("links").
		Where(squirrel.Eq{"link": link}).
		ToSql()
}

func buildError(obj, method, place string, err error) error {
	return fmt.Errorf("%s - %s - %s: %w", obj, method, place, err)
}
