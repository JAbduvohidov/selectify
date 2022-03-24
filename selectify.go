package selectify

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"reflect"
)

type Pool[T comparable] struct {
	*pgxpool.Pool
}

func (p *Pool[T]) SelectMany(query string) (elements []*T) {
	rows, err := p.Query(context.Background(), query)
	if err != nil {
		log.Fatal(err)
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
				log.Fatal(err)
			}

			for i, item := range items {
				value.Elem().Field(i).Set(reflect.ValueOf(*item.(*interface{})))
			}
		default:
			err = rows.Scan(&element)
			if err != nil {
				log.Fatal(err)
			}
		}

		elements = append(elements, &element)
	}

	return
}
