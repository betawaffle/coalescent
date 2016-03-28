package coalescent

import (
	"sync/atomic"
	"unsafe"

	"github.com/hashicorp/go-immutable-radix"
)

var emptyTree = iradix.New()

type Tree struct {
	ptr unsafe.Pointer // *iradix.Tree
}

func (t *Tree) Clone() Tree {
	return Tree{ptr: atomic.LoadPointer(&t.ptr)}
}

func (t Tree) Get() *iradix.Tree {
	return (*iradix.Tree)(t.ptr)
}

func (t Tree) GetOrNew() *iradix.Tree {
	if tree := t.Get(); tree != nil {
		return tree
	}
	return emptyTree
}

func (t *Tree) Load() *iradix.Tree {
	return (*iradix.Tree)(atomic.LoadPointer(&t.ptr))
}

func (t *Tree) Store(tree *iradix.Tree) {
	if tree == nil || tree.Len() == 0 {
		tree = emptyTree // to reduce garbage
	}
	atomic.StorePointer(&t.ptr, unsafe.Pointer(tree))
}
