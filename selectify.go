package selectify

import (
	"context"
	"reflect"

	"github.com/jackc/pgx/v4/pgxpool"
)

// Fielder is a wrapper of the Fields method.
// Implement FieldSetter for improving CPU performance.
type Fielder interface {
	// Fields returns addresses of all fields of the struct.
	Fields() []any
}

type scanner interface {
	Scan(dest ...any) error
}

// SelectMany selects the rows returned by the pool.Query and automatically maps it to your given type.
// It returns a slice of the specified type with an error, if any.
// The elements are nil if no rows are selected.
// err returns the same errors as pool.Query.
func SelectMany[T comparable](pool *pgxpool.Pool, ctx context.Context, query string, args ...any) (elements []*T, err error) {
	rows, err := pool.Query(ctx, query, args...)
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

// SelectRow selects the row returned by the pool.QueryRow and automatically maps it to your given type.
// It returns pointer to the specified type with an error, if any.
func SelectRow[T comparable](pool *pgxpool.Pool, ctx context.Context, query string, args ...any) (*T, error) {
	row := pool.QueryRow(ctx, query, args)
	return scan[T](row)
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
