package elasticthought

import "sync"

type Runnable interface {
	Run(wg *sync.WaitGroup)
}
