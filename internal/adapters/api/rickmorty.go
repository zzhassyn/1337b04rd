package api

import (
	"log/slog"
	"net/http"
	"sync"
	"time"
)

type RickMortyClient struct {
	client *http.Client
	cache  map[int]*cacheEntry
	log    *slog.Logger
	mu     sync.RWMutex
	total  int
}

type cacheEntry struct {
	avatarURL string
	name      string
}

type infoResponse struct {
	Info struct {
		Count int `json:"count"`
	} `json:"info"`
}

type characterInfo struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

func NewRickMortyClient(log *slog.Logger) *RickMortyClient {
	return &RickMortyClient{
		client: &http.Client{Timeout: 10 * time.Second},
		log:    log,
		cache:  make(map[int]*cacheEntry),
	}
}
