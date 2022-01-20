package alac

import (
	"github.com/carterpeel/bobcaygeon/rtsp"
	kyoo "github.com/dirkaholic/kyoo"
	"github.com/dirkaholic/kyoo/job"
	"github.com/sirupsen/logrus"
)

type AudioQueue struct {
	finishedChan chan []byte
	callback     func(data []byte)
	pool         *kyoo.JobQueue
}

func NewAudioQueue(maxDecoders int, callback func(data []byte)) (aq *AudioQueue) {
	aq = &AudioQueue{
		finishedChan: make(chan []byte),
		callback:     callback,
	}

	aq.pool = kyoo.NewJobQueue(maxDecoders)
	aq.pool.Start()

	go func() {
		for d := range aq.finishedChan {
			logrus.Infoln("Executing callback")
			aq.callback(d)
			logrus.Infoln("Callback complete")
		}
	}()

	return aq
}

func (aq *AudioQueue) ProcessSession(session *rtsp.Session) {
	for d := range session.DataChan {
		aq.pool.Submit(&job.FuncExecutorJob{
			Func: func() error {
				decoder, _ := New()
				aq.finishedChan <- decoder.decodeFrame(d)
				logrus.Infoln("Job finished")
				return nil
			},
		})
		logrus.Infoln("Submitted job to queue")
	}
}

func (aq *AudioQueue) SetCallback(f func(data []byte)) {
	aq.callback = f
}
