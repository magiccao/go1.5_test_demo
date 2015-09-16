// code modified from idada
// mail: bg5sbk@gmail.com
package slab2

import (
	"sort"
	"sync/atomic"
	"unsafe"
)

const (
	KB = 1 << 10
	MB = KB << 10
)

type MemPool struct {
	slabClasses    []*memClass // slabs
	chunkSizeOrder []int       // order chunk size
}

func NewMemPool(slabMap map[int]int) *MemPool {
	num := len(slabMap)
	if num == 0 {
		return nil
	}

	index := 0
	order := make([]int, len(slabMap))
	for k, _ := range slabMap {
		order[index] = k
		index++
	}
	sort.Ints(order)

	classes := make([]*memClass, num)
	for i := 0; i < num; i++ {
		items, _ := slabMap[order[i]]
		c := &memClass{maxlen: int32(items)}
		for j := 0; j < items; j++ {
			c.Push(&Mem{Data: make([]byte, order[i])})
		}
		classes[i] = c
	}

	return &MemPool{classes, order}
}

func (p *MemPool) Alloc(size int) *Mem {
	i, num := 0, len(p.chunkSizeOrder)
	for ; i < num; i++ {
		if size <= p.chunkSizeOrder[i] {
			size = p.chunkSizeOrder[i]
			break
		}
	}

	if i < num {
		m := p.slabClasses[i].Pop()
		if m != nil {
			return m
		}
	}
	return &Mem{Data: make([]byte, size)}
}

func (p *MemPool) Free(m *Mem) {
	i, num := 0, len(p.chunkSizeOrder)
	for ; i < num; i++ {
		if cap(m.Data) == p.chunkSizeOrder[i] {
			p.slabClasses[i].Push(m)
			break
		}
	}
}

type memClass struct {
	head   unsafe.Pointer
	length int32
	maxlen int32
}

type Mem struct {
	Data []byte
	next unsafe.Pointer
}

func (class *memClass) Push(item *Mem) {
	if atomic.LoadInt32(&class.length) >= class.maxlen {
		return
	}
	for {
		item.next = atomic.LoadPointer(&class.head)
		if atomic.CompareAndSwapPointer(&class.head, item.next, unsafe.Pointer(item)) {
			atomic.AddInt32(&class.length, 1)
			break
		}
	}
}

func (class *memClass) Pop() *Mem {
	var ptr unsafe.Pointer
	for {
		ptr = atomic.LoadPointer(&class.head)
		if ptr == nil {
			break
		}
		if atomic.CompareAndSwapPointer(&class.head, ptr, ((*Mem)(ptr)).next) {
			atomic.AddInt32(&class.length, -1)
			break
		}
	}
	return (*Mem)(ptr)
}
