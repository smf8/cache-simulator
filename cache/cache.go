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
	Options             *Options
	DataReporter        Reporter
	InstructionReporter Reporter

	NumberOfSets uint64
	OffsetBits   uint64 // # of offset bits
	IndexBits    uint64 // # of set index bits, index in direct mapped and set# in set-associative
	TagBits      uint64 // # of tag bits

	// Tags is a 2d array that represents tag data in cache blocks
	// first dimension is for set # and second dimension is for block # in set
	Tags            []*Set
	InstructionTags []*Set
	Dirty           [][]bool // Dirty is same as Tags, but it only contains a flag
}

func NewCache(options *Options) *Cache {
	c := new(Cache)

	c.Options = options
	c.OffsetBits = uint64(math.Log2(float64(options.BlockSize)))
	c.NumberOfSets = options.CacheSize.DCacheSize / (options.Associativity * options.BlockSize)
	c.IndexBits = uint64(math.Log2(float64(c.NumberOfSets)))
	c.TagBits = AddressSize - c.OffsetBits - c.IndexBits

	c.Tags = make([]*Set, c.NumberOfSets)
	c.InstructionTags = make([]*Set, c.NumberOfSets)
	c.Dirty = make([][]bool, c.NumberOfSets)

	for i := range c.Dirty {
		c.Tags[i] = &Set{list.New(), *c.Options}
		c.InstructionTags[i] = &Set{list.New(), *c.Options}
		c.Dirty[i] = make([]bool, options.Associativity)
	}

	c.DataReporter = Reporter{
		CacheOptions:    c.Options,
		ReplacesCounter: 0,
		MissesCounter:   0,
		AccessesCounter: 0,
	}

	c.InstructionReporter = Reporter{
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

// FlushDirty is to write dirty blocks into memory
func (c *Cache) FlushDirty() {
	for _, set := range c.Tags {
		for _, sb := range set.ToArray() {
			if sb.dirty {
				if Debug {
					fmt.Printf("flushing dirty tag: %d\n", sb.tag)
				}

				c.DataReporter.CopiedWordsCounter += c.Options.BlockSize / WordSize
			}
		}
	}
}

func (c *Cache) HandleRequest(cmd CacheCmd) {
	line := c.ParseCacheRequest(cmd.Address)

	if cmd.Type == DataWriteRef {
		// 1 0x.... data write request
		res := c.handleDataWrite(*line)

		if Debug {
			fmt.Printf("reference to [write] %s hit is %t      %d\n", cmd.Address, res, c.DataReporter.CopiedWordsCounter)
		}
	} else {
		if c.Options.Type == Unified {
			res := c.handleDataRead(*line, cmd.Type)

			if Debug {
				fmt.Printf("reference to [unified-read] %s hit is %t\n", cmd.Address, res)
			}
		} else {
			var res bool
			if cmd.Type == InstructionReadRef {
				res = c.handleInstructionRead(*line)
			} else {
				res = c.handleDataRead(*line, cmd.Type)
			}
			if Debug {
				fmt.Printf("reference to [split-read-%d] %s hit is %t\n", cmd.Type, cmd.Address, res)
			}
		}
	}
}

func (c *Cache) handleDataWrite(cr CacheRequest) bool {
	res, fetched, written := c.Tags[cr.SetNumber].Replace(c.Options.WritePolicy, c.Options.WriteMissPolicy, cr.Tag)

	c.DataReporter.FetchedWordsCounter += fetched
	c.DataReporter.CopiedWordsCounter += written
	c.DataReporter.AccessesCounter++

	if res == ConflictMiss {
		c.DataReporter.MissesCounter++
		c.DataReporter.ReplacesCounter++
	} else if res == CompulsoryMiss {
		c.DataReporter.MissesCounter++
	}

	return res == HIT
}

func (c *Cache) handleDataRead(cr CacheRequest, refType int) bool {
	// checking if cr's TAG is present in cache
	res, fetched, written := c.Tags[cr.SetNumber].CheckTag(cr.Tag)

	c.DataReporter.FetchedWordsCounter += fetched
	c.DataReporter.CopiedWordsCounter += written

	if refType == InstructionReadRef {
		c.InstructionReporter.AccessesCounter++

		if res == ConflictMiss {
			c.InstructionReporter.MissesCounter++
			c.InstructionReporter.ReplacesCounter++
		} else if res == CompulsoryMiss {
			c.InstructionReporter.MissesCounter++
		}
	} else {
		c.DataReporter.AccessesCounter++

		if res == ConflictMiss {
			c.DataReporter.MissesCounter++
			c.DataReporter.ReplacesCounter++
		} else if res == CompulsoryMiss {
			c.DataReporter.MissesCounter++
		}
	}

	return res == HIT
}

func (c *Cache) handleInstructionRead(cr CacheRequest) bool {
	// checking if cr's TAG is present in cache
	res, fetched, _ := c.InstructionTags[cr.SetNumber].CheckTag(cr.Tag)

	c.DataReporter.FetchedWordsCounter += fetched

	c.InstructionReporter.AccessesCounter++

	if res == ConflictMiss {
		c.InstructionReporter.MissesCounter++
		c.InstructionReporter.ReplacesCounter++
	} else if res == CompulsoryMiss {
		c.InstructionReporter.MissesCounter++
	}

	return res == HIT
}
