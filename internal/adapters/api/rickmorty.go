package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

const rickMortyBase = "https://rickandmortyapi.com/api"

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

func (c *RickMortyClient) GetAvatar(ctx context.Context, id int) (avatarURL, name string, err error) {
	c.mu.RLock()
	if entry, ok := c.cache[id]; ok {
		c.mu.RUnlock()

		return entry.avatarURL, entry.name, nil
	}
	c.mu.RUnlock()

	url := fmt.Sprintf("%s/character/%d", rickMortyBase, id)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return "", "", fmt.Errorf("RickMortyClient.GetAvatar: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("RickMortyClient.GetAvatar do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("RickMortyClient.GetAvatar: status %d for id %d", resp.StatusCode, id)
	}

	var info characterInfo

	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", "", fmt.Errorf("RickMortyClient.GetAvatar decode: %w", err)
	}

	c.mu.Lock()
	c.cache[id] = &cacheEntry{avatarURL: info.Image, name: info.Name}
	c.mu.Unlock()

	c.log.Debug("rickmorty: fetched character", "id", id, "name", info.Name)

	return info.Image, info.Name, nil
}

func (c *RickMortyClient) TotalAvatars(ctx context.Context) (int, error) {
	c.mu.RLock()
	if c.total > 0 {
		total := c.total
		c.mu.RUnlock()

		return total, nil
	}
	c.mu.RUnlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rickMortyBase+"/character", http.NoBody)
	if err != nil {
		return 0, fmt.Errorf("RickMortyClient.TotalAvatars: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("RickMortyClient.TotalAvatars do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("RickMortyClient.TotalAvatars: status %d", resp.StatusCode)
	}

	var info infoResponse

	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return 0, fmt.Errorf("RickMortyClient.TotalAvatars decode: %w", err)
	}

	c.mu.Lock()
	c.total = info.Info.Count
	c.mu.Unlock()

	c.log.Debug("rickmorty: total character", "count", info.Info.Count)

	return info.Info.Count, nil
}
