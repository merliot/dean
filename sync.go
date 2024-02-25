//go:build !tinygo

package dean

import (
	//"sync"
	sync "github.com/sasha-s/go-deadlock"
)

type mutex struct {
	sync.Mutex
}

type rwMutex struct {
	sync.RWMutex
}
