package main

import (
	"fmt"
	"math"
	"runtime"
	"strconv"
	"strings"
)

func goroutineID() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}

func main() {
	fmt.Printf("Goroutine ID :: %v\n", goroutineID())
	fmt.Printf("Num Goroutines :: %v\n", runtime.NumGoroutine())

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	oneMillion := math.Pow(10, 6)
	fmt.Printf("Memory Allocated: %.2f MBs | %.2f bytes\n", float64(memStats.Alloc)/oneMillion, float64(memStats.Alloc))
	fmt.Printf("Memory Obtained From Sys: %.2f MBs | %.2f bytes\n", float64(memStats.Sys)/oneMillion, float64(memStats.Sys))
}

// Goroutine ID :: 1
// Num Goroutines :: 1
// Memory Allocated: 0.06 MBs | 62440.00 bytes
// Memory Obtained From Sys: 8.21 MBs | 8211472.00 bytes
