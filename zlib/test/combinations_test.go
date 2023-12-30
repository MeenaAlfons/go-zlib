package test

import (
	"fmt"
	"testing"

	"github.com/MeenaAlfons/go-zlib/zlib/common"
)

func RunTestOnCombinations(
	t *testing.T,
	samples []sample,
	compressOptsList []common.CompressOptions,
	dictionaryFactories []DictionaryFactory,
	repeatCount int,
	combinationPredicate CombinationPredicate,
	decompressOptsFactory func(common.CompressOptions) common.DecompressOptions,
	operation func([]byte, common.CompressOptions, func(common.CompressOptions) common.DecompressOptions, testLogger) error,
	expectedError error,
) {

	for _, sample := range samples {
		for _, compressOpts := range compressOptsList {
			for _, dictionaryFactory := range dictionaryFactories {
				dictionary := dictionaryFactory(sample.decompressed, compressOpts)
				if !combinationPredicate(compressOpts, sample.decompressed, dictionary) {
					continue
				}
				opts := copyCompressOptions(compressOpts)
				opts = opts.WithInitialDictionary(dictionary)

				repeat(repeatCount, func(index int) {
					name := fmt.Sprintf("%s l:%d w:%d h:%d s:%d dict:%d index:%d", sample.name, opts.Level(), opts.WindowBits(), opts.Header(), opts.Strategy(), len(opts.InitialDictionary()), index)
					t.Run(name, func(t *testing.T) {
						err := operation(sample.decompressed, opts, decompressOptsFactory, t)
						if err != expectedError {
							t.Fatalf("Expected error: %v\nActual error: %v", expectedError, err)
						}
					})
				})
			}
		}
	}
}

type CombinationPredicate func(opts common.CompressOptions, sample []byte, dictionary []byte) bool

func And(predicates ...CombinationPredicate) CombinationPredicate {
	return func(opts common.CompressOptions, sample []byte, dictionary []byte) bool {
		for _, predicate := range predicates {
			if !predicate(opts, sample, dictionary) {
				return false
			}
		}
		return true
	}
}

func Not(predicate CombinationPredicate) CombinationPredicate {
	return func(opts common.CompressOptions, sample []byte, dictionary []byte) bool {
		return !predicate(opts, sample, dictionary)
	}
}

func AllCombinationsPredicate(_ common.CompressOptions, _ []byte, _ []byte) bool {
	return true
}

func NilDictionaryPredicate(_ common.CompressOptions, _ []byte, dictionary []byte) bool {
	return dictionary == nil
}

func HeaderTypeRawPredicate(opts common.CompressOptions, _ []byte, _ []byte) bool {
	return opts.Header() == common.HeaderTypeRaw
}
