package manager

import (
	"context"
	"sync"
)

type Manager interface {
	Control(ctx context.Context, wg *sync.WaitGroup)
}
