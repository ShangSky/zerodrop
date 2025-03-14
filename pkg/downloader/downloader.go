package downloader

import (
	"io"
	"sync"
)

type Downloader struct {
	WriterC    chan io.Writer
	Stoped     chan struct{}
	Cancel     chan string
	FileLength int64
	Filename   string
	Cache      []byte
}

type DownloaderStore struct {
	store map[string]Downloader
	m     sync.Mutex
}

func New() *DownloaderStore {
	return &DownloaderStore{store: make(map[string]Downloader)}
}

func (ds *DownloaderStore) Get(id string) (Downloader, bool) {
	ds.m.Lock()
	defer ds.m.Unlock()
	r, ok := ds.store[id]
	return r, ok
}

func (ds *DownloaderStore) Set(id string, downloader Downloader) {
	ds.m.Lock()
	defer ds.m.Unlock()
	ds.store[id] = downloader
}

func (ds *DownloaderStore) Delete(id string) {
	ds.m.Lock()
	defer ds.m.Unlock()
	delete(ds.store, id)
}
