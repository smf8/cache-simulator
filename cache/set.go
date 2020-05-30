package cache

import (
	"container/list"
)

const (
	HIT            = 0
	CompulsoryMiss = 1
	ConflictMiss   = 2
)

type Set struct {
	l             *list.List
	associativity int
}

func (s *Set) CheckTag(tag uint64) int {

	for e := s.l.Front(); e != nil; e = e.Next() {
		if e.Value == tag {
			// tag found in set. move it to the head
			s.l.Remove(e)
			s.l.PushFront(tag)

			return HIT
		}
	}

	// tag was not in cache, fetch it
	var missType int
	if s.l.Len() < s.associativity {
		missType = CompulsoryMiss
	} else {
		missType = ConflictMiss
	}

	s.l.PushFront(tag)

	if missType == ConflictMiss {
		s.l.Remove(s.l.Front())
	}

	return missType
}
