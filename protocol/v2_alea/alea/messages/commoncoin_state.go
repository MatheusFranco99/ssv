package messages

import (
	// "fmt"

	"github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
    "math/rand"
)

/*
 * CommonCoin
 */

type CommonCoin struct {
	Seed int64
	Coin []byte
	Generator *rand.Rand
}

func NewCommonCoin(seed int64) *CommonCoin {

	return &CommonCoin{
		Seed: seed,
		Coin: make([]byte,0),
		Generator: rand.New(rand.NewSource(seed)),
	}
}


func (c *CommonCoin) SetSeed(seed int64)  {
	c.Seed = seed
	c.Generator = rand.New(rand.NewSource(c.Seed))
}

func (c *CommonCoin) HasSeed() bool {
	return c.Seed != 0
}

func (c *CommonCoin) GetCoin(acRound alea.ACRound, round alea.Round) byte {
	size := len(c.Coin)
	wanted_index := int(acRound)*5 + int(round)
	for (size <= wanted_index) {
		coin := c.GenerateCoin()
		c.Coin = append(c.Coin,coin)
		size += 1
	}

	return c.Coin[wanted_index]
}

func (c *CommonCoin) GenerateCoin() byte {
	return byte(c.Generator.Intn(2))
}


