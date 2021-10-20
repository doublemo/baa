package queue

import "testing"

func TestNewPriority(t *testing.T) {
	data := map[int]interface{}{
		10:  "A",
		9:   "B",
		100: "O",
		78:  "U",
		32:  "X",
	}

	pr := NewPriority(data)
	if pr.Pop() != "O" {
		t.Fatal("pop != O")
	}

	pr.Push(1000, "K")
	if pr.Pop() != "K" {
		t.Fatal("pop != K")
	}
}

func TestNewOrderedUint64(t *testing.T) {
	values := []uint64{1, 2, 67, 1000, 90, 8, 3, 4, 6, 90}
	pr := NewOrderedUint64(values...)
	if pr.Pop() != 1 {
		t.Fatal("pop != 1")
	}

	pr.Push(1)
	if pr.Pop() != 1 {
		t.Fatal("pop != 1")
	}

	if pr.Len() != len(values)-1 {
		t.Fatal("len err")
	}

	for pr.Len() > 0 {
		t.Log("pop -> ", pr.Pop())
	}

	t.Log("Len -> ", pr.Len())
}
