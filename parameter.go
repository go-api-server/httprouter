package rest

import (
	"reflect"
)

type parameter struct {
	NotNull bool
	Type    reflect.Type
}
