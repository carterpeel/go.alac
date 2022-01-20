package alac

import (
	"github.com/carterpeel/bobcaygeon/rtsp"
	kyoo "github.com/dirkaholic/kyoo"
	"github.com/dirkaholic/kyoo/job"
)

type AudioQueue struct {
	finishedChan chan []byte
	callback     func(data []byte)
	pool         *kyoo.JobQueue
	decoders     []*Alac
	maxDecoders  int
}

func NewAudioQueue(maxDecoders int, callback func(data []byte)) (aq *AudioQueue) {
	aq = &AudioQueue{
		finishedChan: make(chan []byte),
		callback:     callback,
		decoders:     make([]*Alac, maxDecoders),
		maxDecoders:  maxDecoders,
	}

	aq.pool = kyoo.NewJobQueue(maxDecoders)
	aq.pool.Start()

	for i := range aq.decoders {
		aq.decoders[i], _ = New()
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
		aq.pool.Submit(&job.FuncExecutorJob{
			Func: func() error {
				offset := decoderOffset
				decoder := aq.decoders[offset]
				aq.finishedChan <- decoder.decodeFrame(d)
				return nil
			},
		})
		decoderOffset++
		if decoderOffset >= aq.maxDecoders {
			decoderOffset = 0
		}
	}
}

func (aq *AudioQueue) SetCallback(f func(data []byte)) {
	aq.callback = f
}
