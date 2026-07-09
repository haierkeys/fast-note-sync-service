// Package keyedmutex provides per-key mutual exclusion.
// Package keyedmutex 提供按 key 粒度的互斥锁。
package keyedmutex

import "sync"

// KeyedMutex serializes operations that share the same key while letting operations
// on different keys run concurrently. Unlike singleflight.Group, every caller's
// function actually runs (once its turn comes), it is just never run concurrently
// with another caller holding the same key.
// KeyedMutex 让持有相同 key 的操作互相串行，不同 key 之间互不阻塞。与 singleflight.Group
// 不同，每个调用方的函数都会真正执行一次（等到轮到自己），只是不会与持有相同 key 的其他
// 调用并发执行。
type KeyedMutex struct {
	mu    sync.Mutex
	locks map[string]*refCountedMutex
}

type refCountedMutex struct {
	mu  sync.Mutex
	ref int
}

// New creates an empty KeyedMutex.
// New 创建一个空的 KeyedMutex。
func New() *KeyedMutex {
	return &KeyedMutex{locks: make(map[string]*refCountedMutex)}
}

// Lock acquires the lock for key, blocking until it is available, and returns an
// unlock function that must be called exactly once (typically via defer) to release
// it. Entries are removed once no caller holds or waits on them, so memory usage
// stays bounded by the number of currently-active keys, not the historical total.
// Lock 获取 key 对应的锁，阻塞直至可用，返回的 unlock 函数必须（通常通过 defer）恰好调用
// 一次以释放锁。当没有调用方持有或等待某个 key 时，对应条目会被清理，内存占用只与当前活跃
// key 数量相关，不会随历史调用次数无界增长。
func (m *KeyedMutex) Lock(key string) (unlock func()) {
	m.mu.Lock()
	e, ok := m.locks[key]
	if !ok {
		e = &refCountedMutex{}
		m.locks[key] = e
	}
	e.ref++
	m.mu.Unlock()

	e.mu.Lock()

	return m.unlockFunc(key, e)
}

// TryLock acquires the lock for key only if it is currently free (no other caller
// holds it or is waiting on it), without blocking. ok is false if the key is
// contended; callers should fall back to Lock in that case. This lets a caller that
// holds a snapshot fetched just before calling TryLock treat it as safely fresh only
// when it wins uncontended — if another goroutine is concurrently operating on the
// same key, the snapshot may already be stale, so it must be re-fetched instead.
// TryLock 仅在 key 当前空闲（无人持有或等待）时获取锁，不阻塞；key 被占用时 ok 为 false，
// 调用方应回退到 Lock。这让调用方在 TryLock 无竞争成功时才能信任调用前拿到的快照数据是新鲜
// 的——一旦有其他 goroutine 正在并发操作同一 key，该快照就可能已过期，必须重新查询。
func (m *KeyedMutex) TryLock(key string) (unlock func(), ok bool) {
	m.mu.Lock()
	if _, exists := m.locks[key]; exists {
		m.mu.Unlock()
		return nil, false
	}
	e := &refCountedMutex{ref: 1}
	e.mu.Lock() // uncontended by construction, always succeeds immediately // 刚创建，必定无竞争，立即成功
	m.locks[key] = e
	m.mu.Unlock()

	return m.unlockFunc(key, e), true
}

func (m *KeyedMutex) unlockFunc(key string, e *refCountedMutex) func() {
	var once sync.Once
	return func() {
		once.Do(func() {
			e.mu.Unlock()

			m.mu.Lock()
			e.ref--
			if e.ref == 0 {
				delete(m.locks, key)
			}
			m.mu.Unlock()
		})
	}
}
