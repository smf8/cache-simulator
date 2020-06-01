package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/smf8/cache-simulator/cache"
)

func main() {

	file, err := os.Open("file.trace")
	r := bufio.NewReader(file)
	if err != nil {
		panic(err)
	}

	_, cmds := readInput(r)
	for i := cache.WriteAllocatePolicy; i <= cache.NoWriteAllocatePolicy; i++ {
		for j := cache.WriteBackPolicy; j <= cache.WriteThroughPolicy; j++ {
			for blockSize := uint64(1 << 6); blockSize <= 128; blockSize *= 2 {
				for cacheSize := uint64(1 << 13); cacheSize <= 1<<14; cacheSize *= 2 {
					for associativity := uint64(2); associativity <= 4; associativity *= 2 {
						var hitPolicy, missPolicy string

						if i == cache.WriteBackPolicy {
							hitPolicy = "WB"
						} else {
							hitPolicy = "WT"
						}
						if j == cache.WriteThroughPolicy {
							missPolicy = "WA"
						} else {
							missPolicy = "NW"
						}
						options := &cache.Options{
							Type:            cache.Split,
							BlockSize:       blockSize,
							CacheSize:       cache.CacheSize{cacheSize, cacheSize},
							WriteMissPolicy: cache.Policy(j),
							WritePolicy:     cache.Policy(i),
							Associativity:   associativity,
						}
						c := cache.NewCache(options)

						for _, cmd := range cmds {
							c.HandleRequest(cmd)
						}

						// write whatever that is dirty
						c.FlushDirty()

						dataHitRate := 1 - (float64(c.DataReporter.MissesCounter) / float64(c.DataReporter.AccessesCounter))
						instructionHitRate := 1 - (float64(c.InstructionReporter.MissesCounter) / float64(c.InstructionReporter.AccessesCounter))

						fmt.Printf("[Instruction][%s][%s][BS#%d][CS#%d][AS#%d] -> %.6f\n", hitPolicy, missPolicy, blockSize, cacheSize, associativity, instructionHitRate)
						fmt.Printf("[Instruction][%s][%s][BS#%d][CS#%d][AS#%d] -> %.6f\n\n", hitPolicy, missPolicy, blockSize, cacheSize, associativity, dataHitRate)

					}
				}
			}
			fmt.Println("*************** MP_CHANGE ****************")
		}
		fmt.Println("################# HP_CHANGE ##################")
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
