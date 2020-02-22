package queue

import "errors"

var ErrOutOfCapacity = errors.New("out of capacity")

// 一个生产者 & 一个消费者的情况之下
// 生产者只修改 writeIndex，消费者只修改 readIndex
// 由此不需要使用 atomic 包操作
type RingQueue struct {
	// 环形队列长度
	// 可以写入的长度为 queueLen-1，这样就可以区分 IsEmpty/IsFull
	// IsEmpty ： readIndex = writeIndex
	// IsFull：readIndex + writeIndex = queueLen
	queueLen int32
	// 可以读的位置下标
	readIndex int32
	// 可以写的位置下标
	writeIndex int32
	//队列
	queue []byte
}

func NewRingQueue(capacity int32) *RingQueue {
	return &RingQueue{
		queueLen:   capacity + 1,
		readIndex:  0,
		writeIndex: 0,
		queue:      make([]byte, capacity),
	}
}

// 队列已经使用的空间长度
func (rq *RingQueue) getUsedLen() int32 {
	if rq.writeIndex >= rq.readIndex {
		return rq.writeIndex - rq.readIndex
	}
	return rq.queueLen - rq.readIndex + rq.writeIndex
}

func (rq *RingQueue) getLeftLen() int32 {
	return rq.queueLen - 1 - rq.getUsedLen()
}

// 判断队列是否为空
func (rq *RingQueue) isEmpty() bool {
	return rq.writeIndex == rq.readIndex
}

// 判断队列是否满了
func (rq *RingQueue) isFull() bool {
	return rq.writeIndex+rq.readIndex < rq.queueLen
}

func (rq *RingQueue) Save(buf []byte) error {
	if len(buf) > int(rq.getLeftLen()) {
		return ErrOutOfCapacity
	}

	for index, bit := range buf {
		rq.queue[int(rq.writeIndex)+index] = bit
	}

	rq.writeIndex += int32(len(buf))
	rq.readIndex %= rq.queueLen

	return nil
}

func (rq *RingQueue) Get() []byte {
	if rq.getUsedLen() == 0 {
		return nil
	}

	currentWriteIndex := rq.writeIndex

	var retBytes []byte
	if rq.readIndex < currentWriteIndex {
		retBytes = rq.queue[rq.readIndex:currentWriteIndex]
	} else {
		retBytes = append(rq.queue[rq.readIndex:], rq.queue[:currentWriteIndex]...)
	}

	rq.readIndex = currentWriteIndex

	return retBytes
}
