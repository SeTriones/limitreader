package reader

import (
	"sync"
	"testing"

	log "github.com/Sirupsen/logrus"
)

func TestReader(t *testing.T) {
	cnt := 1000000
	log.Infof("total cnt=%d", cnt)
	r := NewReader(200, 50)
	wg := sync.WaitGroup{}
	consumerCnt := 5
	wg.Add(consumerCnt + 1)
	mu := sync.Mutex{}

	m := make(map[int]int)
	cur := 0

	go func(r *Reader) {
		defer wg.Done()
		for i := 0; i < cnt; i++ {
			r.Put(i)
		}
	}(r)

	for i := 0; i < consumerCnt; i++ {
		go func(r *Reader) {
			defer wg.Done()
			queue := r.GetOutput()
			for values := range queue {
				mu.Lock()
				for _, val := range values {
					m[val.(int)] = 1
				}
				cur = cur + len(values)
				if cur == cnt {
					r.Stop()
				}
				mu.Unlock()
			}
		}(r)
	}

	wg.Wait()
	if len(m) != cur {
		t.Fatalf("digit missing in reader")
	}
	log.Infof("reader test done")
}
