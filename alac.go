// Package alac - Apple Lossless (ALAC) decoder
package alac

import (
	"fmt"
	"sync"
)

// New alac decoder. Sample size 16, 2 chan!
func New() (*Alac, error) {
	a := newAlac(16, 2)
	if a == nil {
		return nil, fmt.Errorf("can't create alac. No idea why, though")
	}
	// TODO: fmtp stuff
	// fmtp: 96 352 0 16 40 10 14 2 255 0 0 44100
	a.setinfo_max_samples_per_frame = 352 // frame_size;
	a.setinfo_7a = 0                      // fmtp[2];
	a.setinfo_sample_size = 16            // sample_size;
	a.setinfo_rice_historymult = 40       // fmtp[4];
	a.setinfo_rice_initialhistory = 10    // fmtp[5];
	a.setinfo_rice_kmodifier = 14         // fmtp[6];
	a.setinfo_7f = 2                      // fmtp[7];
	a.setinfo_80 = 255                    // fmtp[8];
	a.setinfo_82 = 0                      // fmtp[9];
	a.setinfo_86 = 0                      // fmtp[10];
	a.setinfo_8a_rate = 44100             // fmtp[11];

	// Audio Queue
	a.queue = &AudioQueue{
		mu:           sync.Mutex{},
		finishedChan: make(chan []byte, 8192),
	}

	a.allocateBuffers()
	return a, nil
}

/*
Decode decodes a frame.
*/
func (alac *Alac) Decode(f []byte) []byte {
	return alac.decodeFrame(f)
}

/*
BitDepth returns the number of bits per sample and channel.
*/
func (alac *Alac) BitDepth() int {
	return alac.samplesize
}

/*
NumChannels returns the number of channels to decode.
*/
func (alac *Alac) NumChannels() int {
	return int(alac.setinfo_7f)
}

/*
SampleRate returns the SampleRate in samples per second.
*/
func (alac *Alac) SampleRate() int {
	return int(alac.setinfo_8a_rate)
}
