package alac

import (
	"github.com/carterpeel/bobcaygeon/rtsp"
	kyoo "github.com/dirkaholic/kyoo"
	"github.com/dirkaholic/kyoo/job"
	"sync"
)

type AudioQueue struct {
	finishedChan chan []byte
	callback     func(data []byte)
	pool         *kyoo.JobQueue
	workers      []*worker
	maxDecoders  int
}

type worker struct {
	mu      *sync.Mutex
	decoder *Alac
}

func NewAudioQueue(maxDecoders int, callback func(data []byte)) (aq *AudioQueue) {
	aq = &AudioQueue{
		finishedChan: make(chan []byte),
		callback:     callback,
		workers:      make([]*worker, maxDecoders),
		maxDecoders:  maxDecoders,
	}

	aq.pool = kyoo.NewJobQueue(maxDecoders)
	aq.pool.Start()

	for i := range aq.workers {
		aq.workers[i] = &worker{
			mu: &sync.Mutex{},
		}
		aq.workers[i].decoder, _ = New()
	}

	go func() {
		for d := range aq.finishedChan {
			aq.callback(d)
		}
	}()

	return aq
}

func (aq *AudioQueue) ProcessSession(session *rtsp.Session) {
	var decoderOffset int
	for d := range session.DataChan {
		if decoderOffset >= aq.maxDecoders {
			decoderOffset = 0
		}
		aq.pool.Submit(&job.FuncExecutorJob{
			Func: func() error {
				offset := decoderOffset
				wk := aq.workers[offset]
				wk.mu.Lock()
				aq.finishedChan <- wk.decoder.Decode(d)
				wk.mu.Unlock()
				return nil
			},
		})
		decoderOffset++
	}
}

func (aq *AudioQueue) SetCallback(f func(data []byte)) {
	aq.callback = f
}
