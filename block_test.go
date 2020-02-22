package queue

import "testing"

func TestBlock(t *testing.T) {
	b, err := newBlock(1, 0, []byte("qinhanzxc"))
	if err != nil {
		t.Error(err)
		return
	}
	bytes := b.serialize()

	t.Log(b)

	b1, err := newBlockFromBytes(bytes)
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(b1)
}
