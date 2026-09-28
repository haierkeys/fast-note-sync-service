package websocket_router

import (
	"testing"
	"time"
)

func TestSyncDownloadExpiredKeepsLiveConnection(t *testing.T) {
	now := time.Now()
	open := make(chan struct{})
	e := &syncDownloadEntry{UpdatedAt: now.Add(-30 * time.Minute), ConnDone: open}
	if syncDownloadExpired(e, now) {
		t.Fatal("entry of a live connection must survive past the idle TTL while the client processes a page")
	}
	close(open)
	if !syncDownloadExpired(e, now) {
		t.Fatal("entry must be evicted once its connection is closed")
	}
}

func TestSyncDownloadExpiredFallbacks(t *testing.T) {
	now := time.Now()
	if syncDownloadExpired(&syncDownloadEntry{UpdatedAt: now.Add(-5 * time.Minute)}, now) {
		t.Fatal("recently updated entry must not be evicted")
	}
	if !syncDownloadExpired(&syncDownloadEntry{UpdatedAt: now.Add(-11 * time.Minute)}, now) {
		t.Fatal("entry without a connection signal keeps the old idle TTL")
	}
	live := make(chan struct{})
	if !syncDownloadExpired(&syncDownloadEntry{UpdatedAt: now.Add(-25 * time.Hour), ConnDone: live}, now) {
		t.Fatal("max idle cap applies even to live connections")
	}
}

func TestEvictExpiredSyncDownloads(t *testing.T) {
	now := time.Now()
	live := make(chan struct{})
	closed := make(chan struct{})
	close(closed)
	syncDownloadCacheMap.Store("ttl-live_file", &syncDownloadEntry{UpdatedAt: now.Add(-20 * time.Minute), ConnDone: live})
	syncDownloadCacheMap.Store("ttl-closed_file", &syncDownloadEntry{UpdatedAt: now.Add(-20 * time.Minute), ConnDone: closed})
	defer syncDownloadCacheMap.Delete("ttl-live_file")
	defer syncDownloadCacheMap.Delete("ttl-closed_file")
	evictExpiredSyncDownloads(now)
	if _, ok := syncDownloadGet("ttl-live", "file"); !ok {
		t.Fatal("live entry was evicted")
	}
	if _, ok := syncDownloadGet("ttl-closed", "file"); ok {
		t.Fatal("closed entry was kept")
	}
}
