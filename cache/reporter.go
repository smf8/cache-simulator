package cache

import (
	"fmt"
	"math"
	"strings"
)

type Reporter struct {
	CacheOptions *Options

	AccessesCounter uint64
	MissesCounter   uint64
	ReplacesCounter uint64

	FetchedWordsCounter uint64
	CopiedWordsCounter  uint64
}

func (r *Reporter) ReportSettings() string {
	sb := new(strings.Builder)
	sb.WriteString(fmt.Sprintln("***CACHE SETTINGS***"))

	switch r.CacheOptions.Type {
	case Unified:
		sb.WriteString(fmt.Sprintln("Unified I- D-cache"))
		sb.WriteString(fmt.Sprintf("Size: %d\n", r.CacheOptions.CacheSize.DCacheSize))
		break
	case Split:
		sb.WriteString(fmt.Sprintln("Split I- D-cache"))
		sb.WriteString(fmt.Sprintf("I-cache size: %d\n", r.CacheOptions.CacheSize.ICacheSize))
		sb.WriteString(fmt.Sprintf("D-cache size: %d\n", r.CacheOptions.CacheSize.DCacheSize))
		break
	}
	sb.WriteString(fmt.Sprintf("Associativity: %d\n", r.CacheOptions.Associativity))
	sb.WriteString(fmt.Sprintf("Block size: %d\n", r.CacheOptions.BlockSize))
	if r.CacheOptions.WritePolicy == WriteBackPolicy {
		sb.WriteString(fmt.Sprintf("Write policy: WRITE BACK\n"))
	} else {
		sb.WriteString(fmt.Sprintf("Write policy: WRITE THROUGH\n"))
	}
	if r.CacheOptions.WriteMissPolicy == WriteAllocatePolicy {
		sb.WriteString(fmt.Sprintf("Allocation policy: WRITE ALLOCATE\n"))
	} else {
		sb.WriteString(fmt.Sprintf("Allocation policy: WRITE NO ALLOCATE\n"))
	}

	return sb.String()
}
func (r *Reporter) Report(tp string) string {
	sb := new(strings.Builder)
	sb.WriteString(fmt.Sprintf("%s\n", tp))
	sb.WriteString(fmt.Sprintf("accesses: %d\n", r.AccessesCounter))
	sb.WriteString(fmt.Sprintf("misses: %d\n", r.MissesCounter))

	missRate := float64(r.MissesCounter) / float64(r.AccessesCounter)
	missRate *= 10000
	missRate = math.Round(missRate)
	missRate /= 10000

	if r.AccessesCounter != 0 {
		sb.WriteString(fmt.Sprintf("miss rate: %.4f (hit rate %.4f)\n", missRate, 1-missRate))
	} else {
		sb.WriteString(fmt.Sprintf("miss rate: 0.0000 (hit rate 0.0000)\n"))
	}

	sb.WriteString(fmt.Sprintf("replace: %d\n", r.ReplacesCounter))

	return sb.String()
}
func (r *Reporter) ReportMemoryBus() string {
	sb := new(strings.Builder)
	sb.WriteString(fmt.Sprintf("TRAFFIC (in words)\n"))
	sb.WriteString(fmt.Sprintf("demand fetch: %d\n", r.FetchedWordsCounter))
	sb.WriteString(fmt.Sprintf("copies back: %d\n", r.CopiedWordsCounter))

	return sb.String()
}
