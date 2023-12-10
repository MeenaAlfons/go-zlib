package test

import (
	"fmt"
	"math/rand"
)

type sample struct {
	name         string
	decompressed []byte
}

func getDataSamples() (samples []sample) {
	samples = append(samples, getPredefinedSamples()...)
	samples = append(samples, getRandomSamples()...)
	return
}

func getPredefinedSamples() []sample {
	return []sample{
		{
			name:         "empty",
			decompressed: []byte{},
		},
		{
			name:         "hello world",
			decompressed: []byte("hello world"),
		},
	}
}

func getRandomSamples() (samples []sample) {
	lengths := []int{
		1000,
		10 + 1<<10,
		10 + 1<<15,
		10 + 1<<16,
	}
	for _, length := range lengths {
		samples = append(samples, sample{
			name:         fmt.Sprintf("random %d", length),
			decompressed: RandBytes(length),
		})
	}
	return
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandBytes(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return b
}
