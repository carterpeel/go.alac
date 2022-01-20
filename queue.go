package alac

import (
	"fmt"
	"github.com/carterpeel/bobcaygeon/rtsp"
	"github.com/enriquebris/goworkerpool"
	"sync"
)

type AudioQueue struct {
	mu               sync.Mutex
	finishedChan     chan []byte
	maxDecoders      int
	callback         func(data []byte)
	pool             *goworkerpool.Pool
	lastDecoderIndex int32
	decoders         []*decoderWorker
}

type decoderWorker struct {
	a        *Alac
	index    int32
	ch       chan []byte
	curBytes []byte
}

func NewAudioQueue(maxDecoders int, callback func(data []byte)) (aq *AudioQueue, err error) {
	aq = &AudioQueue{
		mu:               sync.Mutex{},
		finishedChan:     make(chan []byte),
		maxDecoders:      maxDecoders,
		callback:         callback,
		decoders:         make([]*decoderWorker, maxDecoders),
		lastDecoderIndex: -1,
	}

	for i, v := range aq.decoders {
		v = &decoderWorker{
			index: int32(i),
			ch:    aq.finishedChan,
		}
		v.ch = aq.finishedChan
		v.a, _ = New()
	}

	if aq.pool, err = goworkerpool.NewPoolWithOptions(goworkerpool.PoolOptions{
		TotalInitialWorkers:          2,
		MaxWorkers:                   uint(maxDecoders),
		MaxOperationsInQueue:         50,
		WaitUntilInitialWorkersAreUp: true,
	}); err != nil {
		return nil, fmt.Errorf("error initializing new worker pool: %w", err)
	}

	aq.pool.SetWorkerFunc(func(data interface{}) bool {
		v := data.(*decoderWorker)
		aq.finishedChan <- v.a.Decode(v.curBytes)
		return true
	})

	go func() {
		for d := range aq.finishedChan {
			aq.callback(d)
		}
	}()

	return aq, nil
}

func (aq *AudioQueue) ProcessSession(session *rtsp.Session) {
	decoderWorkerIndex := -1
	for d := range session.DataChan {
		decoderWorkerIndex++
		aq.decoders[decoderWorkerIndex].curBytes = d
		if err := aq.pool.AddTask(aq.decoders[decoderWorkerIndex]); err != nil {
			panic(err)
		}
	}
}

func (aq *AudioQueue) SetCallback(f func(data []byte)) {
	aq.callback = f
}
