package queue

import (
	"errors"
	"sync/atomic"
	"time"
)

type MultiArray struct {
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

func NewMultiArray(capacity int32) *MultiArray {
	return &MultiArray{
		queueLen:   capacity + 1,
		readIndex:  0,
		writeIndex: 0,
		queue:      make([]byte, capacity),
	}
}

func (multiArray *MultiArray) Save(buf []byte) error {
	// cal the displacement
	blockCount := len(buf) / int(blockDataSize)
	lastBlockSize := len(buf) % int(blockDataSize)
	if lastBlockSize != 0 {
		blockCount++
	}
	displacement := int32(blockCount * int(blockSize))

	// modify the writeIndex
	_, currentWriteIndex, usedLen := multiArray.getUsedLen()
	if usedLen+displacement > multiArray.queueLen-1 {
		return ErrOutOfCapacity
	}

	if !atomic.CompareAndSwapInt32(&multiArray.writeIndex, currentWriteIndex, (currentWriteIndex+displacement)%multiArray.queueLen) {
		return errors.New("failed to swap")
	}

	// save data
	for i := blockCount; i > 0; i-- {
		var dataToSave []byte
		if i == blockCount && lastBlockSize != 0 {
			dataToSave = append(dataToSave, buf[(blockCount-1)*int(blockDataSize):]...)
		} else {
			start := (i - 1) * int(blockDataSize)
			dataToSave = append(dataToSave, buf[start:start+int(blockDataSize)]...)
		}

		b, _ := newBlock(uint16(blockCount), uint16(i-1), dataToSave)

		index := (int(currentWriteIndex) + (i-1)*int(blockSize)) % int(multiArray.queueLen)
		multiArray.writeBlock(index, b)
	}

	return nil
}

func (multiArray *MultiArray) Get() ([][]byte, error) {
	// get block count
	currentReadIndex, _, usedLen := multiArray.getUsedLen()
	if int(usedLen)%int(blockSize) != 0 {
		return nil, errors.New("get error : usedLen is not n * blockSize")
	}
	blockCount := int(usedLen) / int(blockSize)

	// parse all blocks
	blocks := multiArray.getBlocks(currentReadIndex, blockCount)

	return multiArray.parseBlocks(blocks)
}

// getUsedLen return currentReadIndex, currentWriteIndex, usedLen
func (multiArray *MultiArray) getUsedLen() (int32, int32, int32) {
	currentWriteIndex := atomic.LoadInt32(&multiArray.writeIndex)
	currentReadIndex := atomic.LoadInt32(&multiArray.readIndex)
	var usedLen int32
	if currentWriteIndex >= currentReadIndex {
		usedLen = currentWriteIndex - currentReadIndex
	} else {
		usedLen = multiArray.queueLen - currentReadIndex + currentWriteIndex
	}

	return currentReadIndex, currentWriteIndex, usedLen
}

// writeBlock is to write block details to queue
func (multiArray *MultiArray) writeBlock(startIndex int, b *block) {
	for index, bit := range b.serialize() {
		multiArray.queue[startIndex+index] = bit
	}

	multiArray.queue[startIndex] = 1
}

// writeBlock is to get block details from queue
func (multiArray *MultiArray) getBlock(startIndex int) *block {
	var data []byte
	endIndex := startIndex + int(blockSize)
	if endIndex < int(multiArray.queueLen) {
		data = multiArray.queue[startIndex:endIndex]
	} else {
		data = append(data, multiArray.queue[startIndex:]...)
		data = append(data, multiArray.queue[:endIndex-int(multiArray.queueLen)]...)
	}

	b, _ := newBlockFromBytes(data)
	return b
}

// getBlocks is to get a bunch of blocks from a startIndex
func (multiArray *MultiArray) getBlocks(startIndex int32, blockCount int) []*block {
	// parse all blocks
	var blocks []*block
	for i := 1; i <= blockCount; i++ {
		blocks = append(blocks, multiArray.getBlock(int(startIndex)))
		startIndex += int32(blockSize)
	}

	return blocks
}

// parseBlocks is to parse blocks
func (multiArray *MultiArray) parseBlocks(blocks []*block) ([][]byte, error) {
	var ret [][]byte

	for i := 0; i < len(blocks); {
		if !blocks[i].isCompleted(0) {
			time.Sleep(time.Millisecond * 5)
			if !blocks[i].isCompleted(0) {
				i++
				continue
			}
		}

		// find the first block
		start := i
		i += int(blocks[i].blockCount)

		// get data details
		var data []byte
		for _, b := range blocks[start:i] {
			data = append(data, b.data[:b.blockLen]...)
		}

		ret = append(ret, data)
	}

	return ret, nil
}
