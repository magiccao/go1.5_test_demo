package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
	"unsafe"
)

import "./slab2"

var addrs = struct {
	addrs map[unsafe.Pointer]struct{}
	sync.Mutex
}{
	make(map[unsafe.Pointer]struct{}),
	sync.Mutex{},
}

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
					addrs.Lock()
					if _, ok := addrs.addrs[unsafe.Pointer(&mem.Data[0])]; ok {
						fmt.Printf("%dth goroutine: addr %p is using by other goroutine!\n", num, mem.Data)
						fmt.Printf("%dth current addrs %v\n", num, addrs.addrs)
						addrs.Unlock()
						break
					} else {
						addrs.addrs[unsafe.Pointer(&mem.Data[0])] = struct{}{}
					}
					addrs.Unlock()

					mem.Data[0] = byte(num)
					// simulate task running
					time.Sleep(time.Millisecond * time.Duration(rand.Int()%100))

					if mem.Data[0] != byte(num) {
						fmt.Printf("%dth goroutine buffer edited by other goroutine! buffer = %d\n", num, mem.Data[0])
						break
					}

					addrs.Lock()
					delete(addrs.addrs, unsafe.Pointer(&mem.Data[0]))
					addrs.Unlock()
					pool.Free(mem)
				}
			}
			wg.Done()
		}(i)
	}

	wg.Wait()
}
