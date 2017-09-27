package reader

import (
	"errors"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

var (
	ErrQueueFull = errors.New("Queue Full")
)

type Reader struct {
	buf        chan interface{}
	queueSize  int
	bufSize    int
	bufItemCnt int
	interval   int
	wg         sync.WaitGroup
	output     chan []interface{}
	alive      bool
	flag       bool
	mu         sync.Mutex
}

func NewReader(queueSize int, bufSize int, interval int) *Reader {
	r := &Reader{
		buf:        make(chan interface{}, queueSize),
		queueSize:  queueSize,
		bufSize:    bufSize,
		bufItemCnt: 0,
		wg:         sync.WaitGroup{},
		interval:   interval,
		output:     make(chan []interface{}),
		alive:      true,
		flag:       false,
		mu:         sync.Mutex{},
	}
	go r.timedTask()
	return r
}

func (r *Reader) GetOutput() <-chan []interface{} {
	return r.output
}

func (r *Reader) Put(item interface{}) error {
	select {
	case r.buf <- item:
	default:
		return ErrQueueFull
	}
	r.mu.Lock()
	if len(r.buf) == r.bufSize && !r.flag {
		r.flag = true
		r.mu.Unlock()
		go r.genOutput(r.bufSize, 1)
	} else {
		r.mu.Unlock()
	}
	return nil
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
		if itemCnt == 0 {
			r.mu.Unlock()
			continue
		}
		if itemCnt > r.bufSize {
			itemCnt = r.bufSize
		}
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
	log.Infof("timedTask quit")
}

func (r *Reader) Stop() {
	r.alive = false
	close(r.output)
	close(r.buf)
}
