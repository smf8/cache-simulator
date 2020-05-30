package cache

import (
	"container/list"
	"fmt"
	"math"
	"strconv"
)

const AddressSize = 32 //cache address line length in bits

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
	Options  *Options
	Reporter Reporter

	NumberOfSets uint64
	OffsetBits   uint64 // # of offset bits
	IndexBits    uint64 // # of set index bits, index in direct mapped and set# in set-associative
	TagBits      uint64 // # of tag bits

	// Tags is a 2d array that represents tag data in cache blocks
	// first dimension is for set # and second dimension is for block # in set
	Tags  []*Set
	Dirty [][]bool // Dirty is same as Tags, but it only contains a flag
}

func NewCache(options *Options) *Cache {
	c := new(Cache)

	c.Options = options
	c.OffsetBits = uint64(math.Log2(float64(options.BlockSize)))
	c.NumberOfSets = options.CacheSize.DCacheSize / (options.Associativity * options.BlockSize)
	c.IndexBits = uint64(math.Log2(float64(c.NumberOfSets)))
	c.TagBits = AddressSize - c.OffsetBits - c.IndexBits

	c.Tags = make([]*Set, c.NumberOfSets)
	c.Dirty = make([][]bool, c.NumberOfSets)

	for i := range c.Dirty {
		c.Tags[i] = &Set{list.New(), int(c.Options.Associativity)}
		c.Dirty[i] = make([]bool, options.Associativity)
	}

	c.Reporter = Reporter{
		CacheOptions:    c.Options,
		ReplacesCounter: 0,
		MissesCounter:   0,
		AccessesCounter: 0,
	}
	return c
}

// ParseCacheRequest extracts offset, set # and tag from a address line
func (c *Cache) ParseCacheRequest(address string) *CacheRequest {
	//convert address to hex number
	addr, err := strconv.ParseUint(address, 16, 64)
	if err != nil {
		if Debug {
			fmt.Printf("failed to parse address %s: %s\n", address, err.Error())
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
		fmt.Printf("processing %s: offset #%d, set #%d, tag #%d\n", address, offset, index, tag)
	}

	return &CacheRequest{
		Offset:    offset,
		SetNumber: index,
		Tag:       tag,
	}
}

func (c *Cache) HandleRequest(cmd CacheCmd) {
	switch cmd.Type {
	case DataReadRef:
		// 0 0x....   data read request
		line := c.ParseCacheRequest(cmd.Address)
		res := c.handleDataRead(*line)

		if Debug {
			fmt.Printf("reference to [read] %s hit is %t\n", cmd.Address, res)
		}
		break

	case DataWriteRef:
		// 1 0x.... data write request
		break

	case InstructionReadRef:
		// 2 0x... instruction read request
		break
	}
}

func (c *Cache) handleDataRead(cr CacheRequest) bool {
	// checking if cr's TAG is present in cache
	res := c.Tags[cr.SetNumber].CheckTag(cr.Tag)
	c.Reporter.AccessesCounter++
	if res == ConflictMiss {
		c.Reporter.MissesCounter++
		c.Reporter.ReplacesCounter++
	} else if res == CompulsoryMiss {
		c.Reporter.MissesCounter++
	}

	return res == HIT
}
