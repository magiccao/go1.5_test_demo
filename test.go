package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
	"unsafe"
)

import "./slab2"

var addrs = make(map[unsafe.Pointer]struct{})

func main() {
	var wg sync.WaitGroup
	pool := slab2.NewMemPool(map[int]int{
		1: 2,
		2: 2,
	})

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(num int) {
			for {
				mem := pool.Alloc(1)
				if mem != nil {
					if _, ok := addrs[unsafe.Pointer(&mem.Data[0])]; ok {
						fmt.Printf("%dth goroutine: addr %p is using by other goroutine!\n", num, mem.Data)
						fmt.Printf("%dth current addrs %v\n", num, addrs)
						break
					} else {
						addrs[unsafe.Pointer(&mem.Data[0])] = struct{}{}
					}

					mem.Data[0] = byte(num)
					// simulate task running
					time.Sleep(time.Millisecond * time.Duration(rand.Int()%100))

					if mem.Data[0] != byte(num) {
						fmt.Printf("%dth goroutine buffer edited by other goroutine! buffer = %d\n", num, mem.Data[0])
						break
					}

					delete(addrs, unsafe.Pointer(&mem.Data[0]))
					pool.Free(mem)
				}
			}
			wg.Done()
		}(i)
	}

	wg.Wait()
}
