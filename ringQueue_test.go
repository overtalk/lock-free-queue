package queue_test

import (
	"fmt"
	queue "github.com/overtalk/lockFreeQueue"
	"testing"
)

func TestCircleArray(t *testing.T) {
	originStr := "qinhan"

	fmt.Println(len([]byte(originStr)))

	arr := queue.NewRingQueue(10)
	if err := arr.Save([]byte(originStr)); err != nil {
		t.Error(err)
		return
	}

	fmt.Println(string(arr.Get()))
}
