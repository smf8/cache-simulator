package cache

import (
	"fmt"
	"math"
	"strconv"
)

const CacheAddressSize = 32 //cache address line length in bits

type CacheCmd struct {
	Type    int
	Address string
}

type CacheRequest struct {
	Offset    uint64
	SetNumber uint64
	Tag       uint64
}

type Cache struct {
	Options *Options

	NumberOfSets uint64
	OffsetBits   uint64 // # of offset bits
	IndexBits    uint64 // # of set index bits, index in direct mapped and set# in set-associative
	TagBits      uint64 // # of tag bits

	// Tags is a 2d array that represents tag data in cache blocks
	// first dimension is for set # and second dimension is for block # in set
	Tags  [][]uint64
	Dirty [][]bool // Dirty is same as Tags, but it only contains a flag
}

func NewCache(options *Options) *Cache {
	c := new(Cache)

	c.Options = options
	c.OffsetBits = uint64(math.Log2(float64(options.BlockSize)))
	c.NumberOfSets = options.CacheSize.DCacheSize / (options.Associativity * options.BlockSize)
	c.IndexBits = uint64(math.Log2(float64(c.NumberOfSets)))
	c.TagBits = CacheAddressSize - c.OffsetBits - c.IndexBits

	c.Tags = make([][]uint64, c.NumberOfSets)
	c.Dirty = make([][]bool, c.NumberOfSets)

	for i := range c.Tags {
		c.Tags[i] = make([]uint64, options.Associativity)
		c.Dirty[i] = make([]bool, options.Associativity)
	}

	return c
}
func (c *Cache) ParseCacheRequest(address string) *CacheRequest {
	//convert address to hex number
	addr, err := strconv.ParseUint(address, 16, 64)
	if err != nil {
		if Debug {
			fmt.Printf("failed to parse address %s: %s", address, err.Error())
		}

		return nil
	}
	// first cut offset bits from address
	offset := addr & ((1 << c.OffsetBits) - 1)
	// secondly, extract index bits

	addr = addr >> c.OffsetBits
	index := addr & ((1 << c.IndexBits) - 1)
	addr = addr >> c.IndexBits

	// what is left is tag
	tag := addr

	if Debug {
		fmt.Printf("processing %s: offset #%d, set #%d, tag #%d", address, offset, index, tag)
	}

	return &CacheRequest{
		Offset:    offset,
		SetNumber: index,
		Tag:       tag,
	}
}
