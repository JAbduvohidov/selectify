package selectify

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"reflect"
)

type Pool struct {
	*pgxpool.Pool
}

func SelectMany[T comparable](pool *Pool, ctx context.Context, query string) (elements []*T, err error) {
	rows, err := pool.Query(ctx, query)
	if err != nil {
		return
	}

	for rows.Next() {
		var element T

		value := reflect.ValueOf(&element)
		switch value.Elem().Kind() {
		case reflect.Struct:
			var items []interface{}

			for i := 0; i < value.Elem().NumField(); i++ {
				var v = value.Elem().Field(i).Interface()
				items = append(items, &v)
			}

			err = rows.Scan(items...)
			if err != nil {
				return
			}

			for i, item := range items {
				value.Elem().Field(i).Set(reflect.ValueOf(*item.(*interface{})))
			}
		default:
			err = rows.Scan(&element)
			if err != nil {
				return
			}
		}

		elements = append(elements, &element)
	}

	return
}
