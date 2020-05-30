package cache

const (
	Unified = iota
	Split   = iota
)

const (
	WriteBackPolicy       = iota
	WriteThroughPolicy    = iota
	WriteAllocatePolicy   = iota
	NoWriteAllocatePolicy = iota
)

const (
	DataReadRef        = iota
	DataWriteRef       = iota
	InstructionReadRef = iota
)

const WordSize = 4

const Debug = true

type CacheType uint64

type Policy uint64

//Options is options for our cache
//
// word size in cache is 4 bytes
type Options struct {
	BlockSize       uint64 // Cache block size in byte
	Type            CacheType
	Associativity   uint64
	WritePolicy     Policy
	WriteMissPolicy Policy
	CacheSize       CacheSize
}

type CacheSize struct {
	ICacheSize uint64
	DCacheSize uint64
}
