package test

import (
	"github.com/MeenaAlfons/go-zlib/zlib/common"
)

func getCompressOptionsCombinations() []common.CompressOptions {
	windowBits := []int{
		// 8 is intentionally omitted because it is not `well` supported by the zlib library
		9,
		10,
		11,
		12,
		13,
		14,
		15,
	}
	levels := []int{
		-1,
		0,
		1,
		2,
		3,
		4,
		5,
		6,
		7,
		8,
		9,
	}
	header := []common.HeaderType{
		common.HeaderTypeZlib,
		common.HeaderTypeRaw,
	}
	strategy := []common.StrategyType{
		common.StrategyDefault,
		common.StrategyFiltered,
		common.StrategyHuffmanOnly,
		common.StrategyRLE,
		common.StrategyFixed,
	}

	optsList := []common.CompressOptions{}
	for _, w := range windowBits {
		for _, l := range levels {
			for _, h := range header {
				for _, s := range strategy {
					optsList = append(optsList, common.DefaultCompressOptions().WithLevel(l).WithWindowBits(w).WithHeader(h).WithStrategy(s))
				}
			}
		}
	}

	return optsList
}
