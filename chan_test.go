package ringchan

import (
    "testing"
    "time"
)

func TestNewRingChan(t *testing.T) {
    ch := NewRingChan(2, 5)
    ch.OnDrop = func(i interface{}) {
        t.Log("Drop", i)
    }
    go func() {
         for {
             val := <- ch.Out
             t.Log("Read value:", val)
             time.Sleep(time.Second)
         }
    }()

    for i := 0;i<20;i++ {
        ch.In <- i
    }

    time.Sleep(time.Second * 21)
}

func TestNoDropNewRingChan(t *testing.T) {
    ch := NewRingChan(2, 0)
    ch.OnDrop = func(i interface{}) {
        t.Log("Drop", i)
    }
    go func() {
        for {
            val := <- ch.Out
            t.Log("Read value:", val)
            time.Sleep(time.Second)
        }
    }()

    for i := 0;i<20;i++ {
        ch.In <- i
    }

    time.Sleep(time.Second * 21)
}
