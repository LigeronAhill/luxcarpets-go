package utils

import (
	"fmt"

	"github.com/samber/lo"
)

func Wrap(message string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

type Array []any

func (a Array) Map(fn func(item any, index int) any) Array {
	res := lo.Map(a, fn)
	return Array(res)
}
