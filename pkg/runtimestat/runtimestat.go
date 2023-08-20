package runtimestat

import (
	"espresso/pkg/runtimestat/internal"
)

type RuntimeStat = internal.RuntimeStat
type Option = internal.Option

func Address(address string) Option {
	return internal.Address(address)
}

func New(option ...Option) *RuntimeStat {
	return internal.New(option...)
}
