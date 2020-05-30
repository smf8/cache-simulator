package cache

import "container/list"

type Set struct {
	l             list.List
	associativity int
}

func (s *Set) CheckTag(tag uint64) bool {

	for e := s.l.Back(); e.Next() != nil; e = e.Next() {
		if e.Value == tag {
			// tag found in set. move it to the head
			s.l.Remove(e)
			s.l.PushBack(tag)

			return true
		}
	}
	// tag was not in cache, fetch it
	s.l.Remove(s.l.Front())
	s.l.PushBack(tag)

	return false
}
