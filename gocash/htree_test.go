package gocash

import (
	"reflect"
	"testing"
)

func TestHtree(t *testing.T) {
	htree := NewHash()
	a, b, c, one, two, three := []byte("a"), []byte("b"), []byte("c"), []byte("one"), []byte("two"), []byte("three")

	var err error
	err = htree.Put(a, one)
	if err != nil {
		t.Errorf("Put(%v, %v) expected err == nil, got %v", a, one, err)
	}
	htree.Put(b, two)
	if err != nil {
		t.Errorf("Put(%v, %v) expected err == nil, got %v", b, two, err)
	}
	htree.Put(c, three)
	if err != nil {
		t.Errorf("Put(%v, %v) expected err == nil, got %v", c, three, err)
	}

	var got []byte

	got, err = htree.Get(a)
	if err != nil || !reflect.DeepEqual(got, one) {
		t.Errorf("Get(%v) expected %v, got %v with error %v", a, one, got, err)
	}

	got, err = htree.Get(b)
	if err != nil || !reflect.DeepEqual(got, two) {
		t.Errorf("Get(%v) expected %v, got %v with error %v", b, two, got, err)
	}

	got, err = htree.Remove(c)
	if err != nil || !reflect.DeepEqual(got, three) {
		t.Errorf("Remove(%v) expected %v, got %v with error %v", c, three, got, err)
	}

	got, err = htree.Get(c)
	if err == nil {
		t.Errorf("Remove(%v) expected error, got %v with error %v", c, got, err)
	}
}

func TestHtreeSnapshot(t *testing.T) {
	htree := NewHash()
	a, b, c, one, two, three := []byte("a"), []byte("b"), []byte("c"), []byte("one"), []byte("two"), []byte("three")

	var err error
	err = htree.Put(a, one)
	if err != nil {
		t.Errorf("Put(%v, %v) expected err == nil, got %v", a, one, err)
	}

	var got []byte

	got, err = htree.Get(a)
	if err != nil || !reflect.DeepEqual(got, one) {
		t.Errorf("Get(%v) expected %v, got %v with error %v", a, one, got, err)
	}

	htree.Snapshot()

	got, err = htree.Get(a)
	if err != nil || !reflect.DeepEqual(got, one) {
		t.Errorf("Get(%v) expected %v, got %v with error %v", a, one, got, err)
	}

	htree.Put(b, two)
	if err != nil {
		t.Errorf("Put(%v, %v) expected err == nil, got %v", b, two, err)
	}
	htree.Put(c, three)
	if err != nil {
		t.Errorf("Put(%v, %v) expected err == nil, got %v", c, three, err)
	}

	got, err = htree.Remove(c)
	if err != nil || !reflect.DeepEqual(got, three) {
		t.Errorf("Remove(%v) expected %v, got %v with error %v", c, three, got, err)
	}

	got, err = htree.Get(c)
	if err == nil {
		t.Errorf("Remove(%v) expected error, got %v with error %v", c, got, err)
	}

}
