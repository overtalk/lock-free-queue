package queue_test

import (
	"fmt"
	queue "github.com/overtalk/lockFreeQueue"
	"sync"
	"testing"
)

func TestMultiArray(t *testing.T) {
	ma := queue.NewMultiArray(1000)
	wg := &sync.WaitGroup{}
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func(i int) {
			err := ma.Save([]byte(fmt.Sprintf("testtesttesttesttest%d", i)))
			wg.Done()
			if err != nil {
				t.Error(err)
				return
			}
		}(i)
	}

	wg.Wait()
	ret, err := ma.Get()
	if err != nil {
		t.Error(err)
		return
	}

	for k, v := range ret {
		t.Log(k, string(v))
	}

}
