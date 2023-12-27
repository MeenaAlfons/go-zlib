package test

import (
	"github.com/MeenaAlfons/go-zlib/zlib/common"
)

type DictionaryFactory func([]byte, common.CompressOptions) []byte

func getDictionaryFactories() []DictionaryFactory {
	factories := []DictionaryFactory{
		NilDictionaryFactory,
	}

	sources := []string{
		"data",
		"random",
	}
	sizes := []string{
		"lessThanWindow",
		"sameAsWindow",
		"moreThanWindow",
	}
	for _, source := range sources {
		for _, size := range sizes {
			factories = append(factories, getDictionaryFactory(source, size))
		}
	}

	return factories
}

func NilDictionaryFactory(_ []byte, _ common.CompressOptions) []byte {
	return nil
}

func getDictionaryFactory(sourceType, sizeType string) DictionaryFactory {
	return func(sample []byte, opts common.CompressOptions) []byte {
		size := 1 << opts.WindowBits()
		switch sizeType {
		case "lessThanWindow":
			size = size / 2
		case "moreThanWindow":
			size = size * 2
		case "sameAsWindow":
			// size = size
		}

		if size <= 0 {
			return nil
		}

		var dictionary []byte
		switch sourceType {
		case "data":
			dictionary = repeatString(sample, size)
		case "random":
			dictionary = RandBytes(size)
		}

		return dictionary

	}
}

func repeatString(sample []byte, size int) []byte {
	if len(sample) == 0 {
		return nil
	}
	dictionary := make([]byte, size)
	for i := 0; i < size; i += len(sample) {
		copy(dictionary[i:], sample)
	}
	return dictionary
}
