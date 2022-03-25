package selectify

import (
	"context"
	"reflect"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Pool struct {
	*pgxpool.Pool
}

// Fielder is a wrapper of the Fields method.
// Implement FieldSetter for improving CPU performance.
type Fielder interface {
	// Fields returns addresses of all fields of the struct.
	Fields() []any
}

func SelectMany[T comparable](pool *Pool, ctx context.Context, query string) (elements []*T, err error) {
	rows, err := pool.Query(ctx, query)
	if err != nil {
		return
	}

	for rows.Next() {
		element, err := scan[T](rows)
		if err != nil {
			return nil, err
		}
		elements = append(elements, element)
	}
	err = rows.Err()
	return
}

func SelectRow[T comparable](pool *Pool, ctx context.Context, query string) (*T, error) {
	row := pool.QueryRow(ctx, query)
	return scan[T](row)
}

type scanner interface {
	Scan(dest ...any) error
}

func scan[T comparable](row scanner) (*T, error) {
	element := new(T)

	if f, ok := any(element).(Fielder); ok {
		// if element's type implements Fielder, then no need of using reflect package
		err := row.Scan(f.Fields()...)
		if err != nil {
			return nil, err
		}
	} else if value := reflect.ValueOf(element).Elem(); value.Kind() == reflect.Struct {
		nf := value.NumField()
		items := make([]any, 0, nf)

		for i := 0; i < nf; i++ {
			v := value.Field(i).Interface()
			items = append(items, &v)
		}

		err := row.Scan(items...)
		if err != nil {
			return nil, err
		}

		for i, item := range items {
			value.Field(i).Set(reflect.ValueOf(*item.(*interface{})))
		}
	} else {
		err := row.Scan(element)
		if err != nil {
			return nil, err
		}
	}
	return element, nil
}
