package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"slices"

	"github.com/kavehmz/prime"
)

type Run struct {
	Name string
	From int
	To   int
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	runs := make([]Run, 0, 1_000_000)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}
		var name string
		var from, to int
		_, _ = fmt.Sscanf(line, "%s ran %d to %d", &name, &from, &to)
		runs = append(runs, Run{name, from, to})
	}

	// part1(runs)
	part2(runs)
}

func part1(runs []Run) {
	var maxNum int
	for _, run := range runs {
		if run.To > maxNum {
			maxNum = run.To
		}
	}

	primes64 := prime.Primes(uint64(maxNum))
	log.Printf("Found %d primes", len(primes64))
	primes := make([]int, len(primes64))
	for i, p := range primes64 {
		primes[i] = int(p)
	}
	findNearestPrimeIndex := func(num int) int {
		i, _ := slices.BinarySearch(primes, num)
		return i
	}

	sum := make(map[string]int)
	for _, run := range runs {
		primeStartIx := findNearestPrimeIndex(run.From)
		primeEndIx := min(findNearestPrimeIndex(run.To), len(primes)-1)
		for primes[primeEndIx] > run.To {
			primeEndIx--
		}
		sum[run.Name] += primeEndIx - primeStartIx + 1
	}

	log.Printf("Found %d sums", len(sum))

	var maxSum int
	var maxName string
	for name, s := range sum {
		if s > maxSum {
			maxSum = s
			maxName = name
		}
	}

	fmt.Printf("%s %d\n", maxName, maxSum)
}

func part2(runs []Run) {
	for i := range runs {
		runs[i].To *= 138739
	}
	part1(runs)
}
