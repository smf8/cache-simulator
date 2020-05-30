package main

import (
	"bufio"
	"fmt"
	"github.com/smf8/cache-simulator/cache"
	"io"
	"os"
	"strconv"
	"strings"
)

func main() {
	options, cmds := readInput(os.Stdin)
	cache := cache.NewCache(options)

	fmt.Println(cmds)
	fmt.Println(cache)
}

func readInput(reader io.Reader) (*cache.Options, []cache.CacheCmd) {
	var err error

	options := new(cache.Options)
	cmds := make([]cache.CacheCmd, 0)

	scanner := bufio.NewScanner(reader)
	scanner.Scan()
	firstLine := scanner.Text()
	optionsString := strings.Split(firstLine, "-")

	options.BlockSize, err = strconv.ParseUint(strings.TrimSpace(optionsString[0]), 10, 64)
	if err != nil && cache.Debug {
		panic(err)
	}
	unified, err := strconv.ParseUint(strings.TrimSpace(optionsString[1]), 10, 64)
	if err != nil && cache.Debug {
		panic(err)
	}

	if unified == 0 {
		options.Type = cache.Unified
	} else {
		options.Type = cache.Split
	}

	options.Associativity, err = strconv.ParseUint(strings.TrimSpace(optionsString[2]), 10, 64)
	if err != nil && cache.Debug {
		panic(err)
	}

	wPolicy := strings.TrimSpace(optionsString[3])
	if wPolicy == "wb" {
		options.WritePolicy = cache.WriteBackPolicy
	} else if wPolicy == "wt" {
		options.WritePolicy = cache.WriteThroughPolicy
	} else if cache.Debug {
		fmt.Printf("Error reading write policy\n")
	}

	aPolicy := strings.TrimSpace(optionsString[4])
	if aPolicy == "wa" {
		options.WriteMissPolicy = cache.WriteAllocatePolicy
	} else if aPolicy == "nw" {
		options.WriteMissPolicy = cache.NoWriteAllocatePolicy
	} else if cache.Debug {
		fmt.Printf("Error reading write allocation policy\n")
	}

	cacheSize := new(cache.CacheSize)
	scanner.Scan()

	if unified == 1 {
		line := strings.Split(scanner.Text(), "-")
		cacheSize.ICacheSize, _ = strconv.ParseUint(strings.TrimSpace(line[0]), 10, 64)
		cacheSize.DCacheSize, _ = strconv.ParseUint(strings.TrimSpace(line[1]), 10, 64)
	} else {
		cacheSize.DCacheSize, _ = strconv.ParseUint(strings.TrimSpace(scanner.Text()), 10, 64)
	}

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

	return options, cmds
}
