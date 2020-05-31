package cache

import (
	"container/list"
	"fmt"
)

const (
	HIT            = 0
	CompulsoryMiss = 1
	ConflictMiss   = 2
)

type Set struct {
	l *list.List
	o Options
}

type SetBlock struct {
	tag   uint64
	dirty bool
}

// lookup is to see if a tag is present in any of blocks in current set
// returns if it existed and a list element representing it's SetBlock
func (s *Set) lookup(tag uint64) (bool, *list.Element) {
	for e := s.l.Front(); e != nil; e = e.Next() {
		sb := e.Value.(*SetBlock)
		if sb.tag == tag {
			return true, e
		}
	}

	return false, nil
}

// CheckTag is for reading data from cache.
// it uses LRU replacement policy
//
// returns MISS/HIT and number of words fetched, written
func (s *Set) CheckTag(tag uint64) (int, uint64, uint64) {

	if hit, e := s.lookup(tag); hit {
		// tag found in set. move it to the head
		sb := e.Value.(*SetBlock)

		s.l.Remove(e)
		s.l.PushFront(sb)

		// when data is present in cache,
		// nothing is laoded from memory
		return HIT, 0, 0
	}

	// tag was not in cache, fetch it
	return s.replace(tag)
}

// Replace is for writing data into memory using different policies.
//
// returns # of words fetched, sent between cache and memory
func (s *Set) Replace(hitPolicy, missPolicy Policy, tag uint64) (int, uint64, uint64) {
	if hitPolicy == WriteBackPolicy {
		if hit, e := s.lookup(tag); hit {
			sb := e.Value.(*SetBlock)
			sb.dirty = true
			// changing LRU priority
			s.l.Remove(e)
			s.l.PushFront(sb)

			if Debug {
				fmt.Printf("Replace %d WB hit\n", tag)
			}

			return HIT, 0, 0
		} else {
			if missPolicy == WriteAllocatePolicy {
				// write miss. go fetch block from memory and update it in cache
				// also write LRU block if it is dirty
				hit, fetched, written := s.replace(tag)
				// since we just updated this block (only in cache):
				s.l.Front().Value.(*SetBlock).dirty = true

				if Debug {
					fmt.Printf("Replace %d WB WA miss\n", tag)
				}

				return hit, fetched, written
			} else {
				// write miss with no write allocate
				// do not touch cache, just update memory directly by 1 word
				// CompulsoryMiss does not play any role so we use it if there is no interacction with cache

				if Debug {
					fmt.Printf("Replace %d WB NWA miss\n", tag)
				}

				return CompulsoryMiss, 0, 1
			}
		}
	} else {
		// Write through policy
		if hit, e := s.lookup(tag); hit {
			// just LRU is updated for block
			// a word is updated in memory
			sb := e.Value.(*SetBlock)
			// changing LRU priority
			s.l.Remove(e)
			s.l.PushFront(sb)
			// don't read anything, just write to memory
			if Debug {
				fmt.Printf("Replace %d WT HIT\n", tag)
			}

			return HIT, 0, 1
		} else {
			if missPolicy == WriteAllocatePolicy {
				// write miss. go fetch block from memory
				// in this policy, we write a single word
				// but we fetch the whole block
				missType, fetched, _ := s.replace(tag)

				if Debug {
					fmt.Printf("Replace %d WT WA miss\n", tag)
				}

				return missType, fetched, 1
			} else {
				// write miss with no write allocate
				// do not touch cache, just update memory directly by 1 word
				// CompulsoryMiss does not play any role so we use it if there is no interacction with cache
				if Debug {
					fmt.Printf("Replace %d WT NWA miss\n", tag)
				}

				return CompulsoryMiss, 0, 1
			}
		}
	}

}

// replace is to add a tag to this set.
// first we check if it can be added without removing any block
// if that is not possible, we remove LRU block and saving it in memory if it is dirty
//
// returns type of miss, words fetched, words written
func (s *Set) replace(tag uint64) (int, uint64, uint64) {

	if s.l.Len() < int(s.o.Associativity) {
		// we have empty block, just fetch
		s.l.PushFront(&SetBlock{tag, false})

		return CompulsoryMiss, s.o.BlockSize / WordSize, 0
	}

	e := s.l.Back()
	lruBlock := e.Value.(*SetBlock)
	if lruBlock.dirty {
		// block is dirty, writing to memory
		if Debug {
			fmt.Printf("dirty block replacement : %d", lruBlock.tag)
		}
		return ConflictMiss, s.o.BlockSize / WordSize, s.o.BlockSize / WordSize
	}
	// fetching but not writing
	s.l.Remove(e)
	s.l.PushFront(&SetBlock{tag, false})

	return ConflictMiss, s.o.BlockSize / WordSize, 0
}

func (s *Set) ToArray() []SetBlock {
	res := make([]SetBlock, s.l.Len())
	i := 0
	for e := s.l.Front(); e != nil; e = e.Next() {
		sb := e.Value.(*SetBlock)
		res[i] = *sb
	}

	return res
}
