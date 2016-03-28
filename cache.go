package coalescent

import "sync"

type Cache struct {
	mu   sync.Mutex
	tree Tree
}

func (c *Cache) Clone() *Cache {
	if c == nil {
		return nil
	}
	return &Cache{tree: c.tree.Clone()}
}

// Delete will unconditionally delete the entry under the given key, returning
// the entry if any was present.
func (c *Cache) Delete(k []byte) (interface{}, bool) {
	c.mu.Lock()
	v, ok := c.delete(k)
	c.mu.Unlock()
	return v, ok
}

// DeleteIf will delete the entry under the given key if it matches the passed
// predicate function, returning true if an entry was deleted. The predicate
// function must not panic, and should be as fast as possible.
func (c *Cache) DeleteIf(k []byte, pred func(interface{}) bool) bool {
	c.mu.Lock()
	ok := c.deleteIf(k, pred)
	c.mu.Unlock()
	return ok
}

// Fetch will find or create an entry under the given key. The passed init
// funcion should be simple, quick, and must not panic. If init is nil, a
// default function will be used. The second return value will be false if a
// new entry was created.
func (c *Cache) Fetch(k []byte, init func() interface{}) (interface{}, bool) {
	if init == nil {
		// FIXME: Is there something sensible we can do here?
		panic("nil init function")
	}
	if v, ok := c.getOrLock(k); ok {
		return v, true
	}
	v := init()
	c.insert(k, v)
	c.mu.Unlock()
	return v, false
}

// Get will return the value, if any, under a given key.
func (c *Cache) Get(k []byte) (interface{}, bool) {
	if tree := c.tree.Load(); tree != nil {
		v, ok := tree.Get(k)
		if ok {
			return v, true
		}
	}
	return nil, false
}

// Insert will unconditionally insert a new entry under the given key,
// returning the old entry if any was present.
func (c *Cache) Insert(k []byte, v interface{}) (interface{}, bool) {
	c.mu.Lock()
	v, ok := c.insert(k, v)
	c.mu.Unlock()
	return v, ok
}

type predicate func(interface{}, bool) bool

// delete must only be called while the lock is held.
func (c *Cache) delete(k []byte) (interface{}, bool) {
	if tree := c.tree.Get(); tree != nil {
		tree, old, ok := tree.Delete(k)
		if ok {
			c.tree.Store(tree)
		}
		return old, ok
	}
	return nil, false
}

// deleteIf must only be called while the lock is held.
func (c *Cache) deleteIf(k []byte, pred func(interface{}) bool) bool {
	if pred == nil {
		return false
	}
	if tree := c.tree.Get(); tree != nil {
		tree, old, ok := tree.Delete(k)
		if ok && pred(old) {
			c.tree.Store(tree)
			return true
		}
	}
	return false
}

// getOrLock does a double-checked Get on the underlying tree, to avoid locking
// if the cache already has an entry for the provided key. It will either
// return the found value and true with the lock released, or nil and false
// with the lock held.
func (c *Cache) getOrLock(k []byte) (interface{}, bool) {
	if tree := c.tree.Load(); tree != nil {
		v, ok := tree.Get(k)
		if ok {
			return v, true
		}
		c.mu.Lock()

		// If the tree hasn't changed, we don't need to recheck.
		if tree == c.tree.Get() {
			return nil, false
		}
	} else {
		c.mu.Lock()
	}
	if tree := c.tree.Get(); tree != nil {
		v, ok := tree.Get(k)
		if ok {
			c.mu.Unlock()
			return v, true
		}
	}
	return nil, false
}

// insert must only be called while the lock is held.
func (c *Cache) insert(k []byte, v interface{}) (interface{}, bool) {
	tree, old, ok := c.tree.GetOrNew().Insert(k, v)
	c.tree.Store(tree)
	return old, ok
}

// insertOrReplaceIf must only be called while the lock is held.
// func (c *Cache) insertOrReplaceIf(k []byte, v interface{}, pred func(interface{}) bool) bool {
// 	if pred == nil {
// 		return false
// 	}
// 	tree, old, ok := c.tree.GetOrNew().Insert(k, v)
// 	if !ok || pred(old) {
// 		c.tree.Store(tree)
// 		return true
// 	}
// 	return false
// }
