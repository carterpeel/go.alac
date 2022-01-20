package alac

import (
	"github.com/carterpeel/bobcaygeon/rtsp"
	"sync"
)

type AudioQueue struct {
	mu           sync.Mutex
	finishedChan chan *finishedJob
	maxDecoders  int
	alacs        []worker
	callback     func(data []byte)
}

type worker struct {
	mu     *sync.Mutex
	active bool
	a      *Alac
	ch     chan *finishedJob
}

type finishedJob struct {
	position int
	data     []byte
}

func NewAudioQueue(maxDecoders int, callback func(data []byte)) *AudioQueue {
	aq := &AudioQueue{
		mu:           sync.Mutex{},
		finishedChan: make(chan *finishedJob, 8192),
		maxDecoders:  maxDecoders,
		alacs:        make([]worker, maxDecoders),
		callback:     callback,
	}
	for i := range aq.alacs {
		aq.alacs[i].a, _ = New()
		aq.alacs[i].mu = &sync.Mutex{}
		aq.alacs[i].ch = aq.finishedChan
	}

	go func() {
		for f := range aq.finishedChan {
			// TODO ensure returned data is ordered
			aq.callback(f.data)
		}
	}()

	return aq
}

func (aq *AudioQueue) ProcessSession(session *rtsp.Session) {
	aq.mu.Lock()
	go func() {
		defer aq.mu.Unlock()
		var position int
		lastWorker := -1
		for d := range session.DataChan {
			position++
		Top:
			if aq.alacs[lastWorker+1].active {
				lastWorker++
				goto Top
			}
			aq.alacs[lastWorker+1].SetJob(d, position)
		}
	}()
}

func (aq *AudioQueue) SetCallback(f func(data []byte)) {
	aq.callback = f
}

func (w *worker) IsActive() bool {
	return w.active
}

func (w *worker) SetJob(data []byte, position int) {
	w.mu.Lock()
	w.active = true
	go func(pos int, d []byte) {
		defer func() {
			w.active = false
			w.mu.Unlock()
		}()
		w.ch <- &finishedJob{
			position: pos,
			data:     w.a.Decode(d),
		}
	}(position, data)
}
