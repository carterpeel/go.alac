package alac

import (
	"math/bits"
	"testing"
)

func BenchmarkLeadingZeroesMathBits_1it(b *testing.B) {
	for i := 0; i < b.N; i++ {
		bits.LeadingZeros32(1)
	}
}
func BenchmarkLeadingZeroesDecoder_1it(b *testing.B) {
	for i := 0; i < b.N; i++ {
		count_leading_zeros(1)
	}
}

func BenchmarkLeadingZeroesMathBits_64it(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for i2 := 0; i2 < 64; i2++ {
			bits.LeadingZeros(1)
		}
	}
}

func BenchmarkLeadingZeroesDecoder_64it(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for i2 := 0; i2 < 64; i2++ {
			count_leading_zeros(1)
		}
	}
}