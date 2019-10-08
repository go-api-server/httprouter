package rest

import (
	"reflect"
)

type parameter struct {
	NotNull bool
	Field   string
	Type    reflect.Type
}
