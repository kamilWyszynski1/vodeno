package client

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

const (
	tableName = "client" // client table name.

	duplicateErrorCode = "23505"
)

var (
	psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	ErrDuplicate = errors.New("entry with given payload already exists")
)

// repo is postgresql implementation of Repository.
type repo struct {
	db *sqlx.DB
}

// NewRepo creates new instance of repo.
func NewRepo(db *sqlx.DB) *repo {
	return &repo{db: db}
}

func (r repo) Insert(ctx context.Context, c Entry) error {
	q := psql.Insert(tableName).
		Columns("email", "title", "content", "mailing_id", "insert_time").
		Values(c.Email, c.Title, c.Content, c.MailingID, c.InsertTime)

	query, args, err := q.ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		pqErr, ok := err.(*pq.Error)
		if ok {
			if pqErr.Code == duplicateErrorCode {
				return ErrDuplicate
			}
		}
		return err
	}
	return nil
}

func (r repo) Delete(ctx context.Context, id int) error {
	q := psql.Delete(tableName).Where(sq.Eq{"id": id})
	query, args, err := q.ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

func (r repo) BatchDelete(ctx context.Context, ids []int) error {
	q := psql.Delete(tableName).Where(sq.Eq{"id": ids})
	query, args, err := q.ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

func (r repo) GetFilter(ctx context.Context, params *getParams) ([]Entry, error) {
	q := psql.Select("*").From(tableName).OrderBy("id")
	if params != nil { // set filtering.
		if params.limit != nil {
			q = q.Limit(uint64(*params.limit))
		}
		if params.offset != nil {
			// we use where id > offset pagination. With create index it will be fast
			// enough even with large amount of data.
			// Another possible way of doing it is builtin postgresql cursor but it's overkill here.
			// https://www.postgresql.org/docs/9.2/plpgsql-cursors.html
			q = q.Where(sq.Gt{"id": *params.offset})
		}
		if params.mailingID != nil {
			q = q.Where(sq.Eq{"mailing_id": *params.mailingID})
		}
	}
	query, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}
	var clients []Entry
	if err := r.db.SelectContext(ctx, &clients, query, args...); err != nil {
		return nil, err
	}
	return clients, nil
}

func (r repo) Get(ctx context.Context, id int) (*Entry, error) {
	q := psql.Select("*").From(tableName).Where(sq.Eq{"id": id})
	query, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}
	var client Entry
	if err := r.db.GetContext(ctx, &client, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &client, nil
}
