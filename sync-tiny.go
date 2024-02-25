//go:build tinygo

package dean

import (
	"sync"
)

type mutex struct {
	sync.Mutex
}

type rwMutex struct {
	sync.RWMutex
}
