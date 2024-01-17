package main

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/kavehmz/prime"
)

type Run struct {
	Name string
	From int
	To   int
}

func main() {
	in, _ := io.ReadAll(os.Stdin)
	lines := strings.Split(string(in), "\n")

	runs := make([]Run, 0, len(lines))
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		var name string
		var from, to int
		_, _ = fmt.Sscanf(line, "%s ran %d to %d", &name, &from, &to)
		runs = append(runs, Run{name, from, to})
	}

	var maxNum int
	for _, run := range runs {
		if run.To > maxNum {
			maxNum = run.To
		}
	}

	primes64 := prime.Primes(uint64(maxNum))
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
		var s int
		for i := primeStartIx; i < len(primes) && primes[i] <= run.To; i++ {
			s++
		}
		sum[run.Name] += s
	}

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
