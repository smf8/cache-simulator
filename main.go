package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"code.cloudfoundry.org/bytefmt"
	"github.com/smf8/cache-simulator/cache"
)

func main() {

	file, err := os.Open("file.trace")
	r := bufio.NewReader(file)
	if err != nil {
		panic(err)
	}

	_, cmds := readInput(r)
	blockSize := uint64(4)
	cacheSize := uint64(1 << 13)
	for {
		if blockSize > (1 << 12) {
			break
		}
		options := &cache.Options{
			Type:            cache.Split,
			BlockSize:       blockSize,
			CacheSize:       cache.CacheSize{cacheSize, cacheSize},
			WriteMissPolicy: cache.WriteAllocatePolicy,
			WritePolicy:     cache.WriteBackPolicy,
			Associativity:   2,
		}
		c := cache.NewCache(options)

		for _, cmd := range cmds {
			c.HandleRequest(cmd)
		}

		// write whatever that is dirty
		c.FlushDirty()

		dataHitRate := 1 - (float64(c.DataReporter.MissesCounter) / float64(c.DataReporter.AccessesCounter))
		instructionHitRate := 1 - (float64(c.InstructionReporter.MissesCounter) / float64(c.InstructionReporter.AccessesCounter))

		fmt.Printf("[Instruction][%s] -> %.6f\n", bytefmt.ByteSize(blockSize), instructionHitRate)
		fmt.Printf("[Data][%s] -> %.6f\n\n", bytefmt.ByteSize(blockSize), dataHitRate)

		//<- time.After(time.Microsecond * 200)

		blockSize *= 2
	}

}

func readInput(reader io.Reader) (*cache.Options, []cache.CacheCmd) {

	cmds := make([]cache.CacheCmd, 0)

	scanner := bufio.NewScanner(reader)
	scanner.Scan()

	for scanner.Scan() {
		lines := scanner.Text()
		line := strings.Split(lines, " ")
		t, _ := strconv.Atoi(strings.TrimSpace(line[0]))
		cmd := cache.CacheCmd{
			Type:    t,
			Address: strings.TrimSpace(line[1]),
		}

		cmds = append(cmds, cmd)
	}

	return nil, cmds
}
