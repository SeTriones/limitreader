package reader

import (
	"sync"
	"time"
)

type Reader struct {
	buf        chan interface{}
	bufSize    int
	bufItemCnt int
	interval   int
	wg         sync.WaitGroup
	output     chan []interface{}
	alive      bool
	flag       bool
	mu         sync.Mutex
}

func NewReader(size int, interval int) *Reader {
	r := &Reader{
		buf:        make(chan interface{}, size),
		bufSize:    size,
		bufItemCnt: 0,
		wg:         sync.WaitGroup{},
		interval:   interval,
		output:     make(chan []interface{}),
		alive:      true,
		flag:       false,
		mu:         sync.Mutex{},
	}
	return r
}

func (r *Reader) GetOutput() <-chan []interface{} {
	return r.output
}

func (r *Reader) Put(item interface{}) {
	r.buf <- item
	r.mu.Lock()
	if len(r.buf) == r.bufSize && !r.flag {
		r.flag = true
		r.mu.Unlock()
		go r.genOutput(r.bufSize, 1)
	} else {
		r.mu.Unlock()
	}
}

func (r *Reader) genOutput(cnt int, from int) {
	defer func() {
		r.flag = false
	}()
	if from == 1 {
		r.wg.Add(1)
		defer r.wg.Done()
	}

	output := make([]interface{}, cnt, cnt)
	for i := 0; i < cnt; i++ {
		output[i] = <-r.buf
	}

	r.output <- output
}

func (r *Reader) timedTask() {
	for r.alive {
		time.Sleep(time.Millisecond * time.Duration(r.interval))
		flag := false
		itemCnt := 0
		r.mu.Lock()
		flag = r.flag
		itemCnt = len(r.buf)
		if !r.flag {
			r.flag = true
		}
		r.mu.Unlock()
		if !flag {
			go r.genOutput(itemCnt, 0)
		} else {
			r.wg.Wait()
		}
	}
}

func (r *Reader) Stop() {
	r.alive = false
}
