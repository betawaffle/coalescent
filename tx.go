package coalescent

import "github.com/hashicorp/go-immutable-radix"

type ReaderTx interface {
	Get(k []byte) (interface{}, bool)
	Root() *iradix.Node
}

type WriterTx interface {
	ReaderTx
	Delete(k []byte) (interface{}, bool)
	Insert(k []byte, v interface{}) (interface{}, bool)
}

// Update will start a new writable transaction and pass it to the provided
// function. The function can make any modifications, and should return true
// if it wishes those modifications to be stored. Only one writable transaction
// will run at a time, but it will not block readers.
func (c *Cache) Update(fn func(WriterTx) bool) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	tx := c.tree.GetOrNew().Txn()
	ok := fn(tx)
	if ok {
		c.tree.Store(tx.Commit())
	}
	return ok
}

// View will call the provided function with a read-only snapshot of the cache.
func (c *Cache) View(fn func(ReaderTx)) {
	fn(c.tree.GetOrNew())
}
