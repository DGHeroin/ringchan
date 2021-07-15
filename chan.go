package ringchan

import (
    "container/list"
    "sync"
)

type (
    RingChan struct {
        In      chan<- interface{}
        Out     <-chan interface{}
        OnDrop  func(interface{})
        mu      sync.RWMutex
        buffer  list.List
        maxSize int
    }
)

func NewRingChan(initSize int, maxSize int) *RingChan {
    r := &RingChan{
        maxSize: maxSize,
    }
    in := make(chan interface{}, initSize)
    out := make(chan interface{}, initSize)
    r.In = in
    r.Out = out
    go r.process(in, out)
    return r
}

func (ch *RingChan) Len() int {
    return len(ch.In) + len(ch.Out) + ch.bufferLen()
}

func (ch *RingChan) process(in, out chan interface{}) {
    defer close(out)
loop:
    for {
        val, ok := <-in
        if !ok {
            break loop
        }
        select {
        case out <- val: // out not full
            continue
        default: // out is full
        }
        ch.bufferPush(val)

        for ch.bufferLen() > 0 {
            select {
            case v2, ok2 := <-in:
                if !ok2 {
                    break loop
                }
                ch.bufferPush(v2)
            case out <- ch.bufferPeek():
                ch.bufferPop()
            }
        }
    }
    // drain
    for ch.bufferLen() > 0 {
        out <- ch.bufferPop()
    }
}

func (ch *RingChan) bufferLen() int {
    ch.mu.RLock()
    defer ch.mu.RUnlock()
    return ch.buffer.Len()
}
func (ch *RingChan) bufferPush(val interface{}) {
    ch.mu.Lock()
    defer ch.mu.Unlock()
    ch.buffer.PushBack(val)
    if ch.maxSize > 0 && ch.buffer.Len() > ch.maxSize {
        front := ch.buffer.Front()
        if ch.OnDrop != nil {
            ch.OnDrop(front.Value)
        }
        ch.buffer.Remove(front)
    }
}
func (ch *RingChan) bufferPeek() interface{} {
    ch.mu.Lock()
    defer ch.mu.Unlock()
    if ch.buffer.Len() == 0 {
        return nil
    }
    return ch.buffer.Front().Value
}

func (ch *RingChan) bufferPop() interface{} {
    ch.mu.Lock()
    defer ch.mu.Unlock()
    if ch.buffer.Len() == 0 {
        return nil
    }
    return ch.buffer.Remove(ch.buffer.Front())
}
