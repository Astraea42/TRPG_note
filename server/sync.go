package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// HFSync syncs data to/from a Hugging Face Dataset for persistence.
// Hugging Face Spaces have ephemeral storage, so we need to persist
// the SQLite DB and uploaded assets to a Dataset repo.
type HFSync struct {
	token    string
	repo     string
	dataDir  string
	dbPath   string
	assetDir string
	client   *http.Client
	lastSync map[string]time.Time
}

func NewHFSync(dbPath string) *HFSync {
	repo := os.Getenv("HF_DATASET_REPO")
	if repo == "" {
		return nil
	}
	// Prefer explicit sync token, fall back to auto-injected HF_TOKEN
	token := os.Getenv("HF_SYNC_TOKEN")
	if token == "" {
		token = os.Getenv("HF_TOKEN")
	}
	if token == "" {
		log.Println("[sync] HF_DATASET_REPO is set but no HF_TOKEN/HF_SYNC_TOKEN found")
		return nil
	}
	return &HFSync{
		token:    token,
		repo:     repo,
		dataDir:  filepath.Dir(dbPath),
		dbPath:   dbPath,
		assetDir: filepath.Join(filepath.Dir(dbPath), "graph_assets"),
		client:   &http.Client{Timeout: 30 * time.Second},
		lastSync: make(map[string]time.Time),
	}
}

// --- API calls ---

func (s *HFSync) downloadFile(path string) ([]byte, error) {
	url := fmt.Sprintf("https://huggingface.co/datasets/%s/raw/main/%s", s.repo, path)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+s.token)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		return nil, nil // file not yet in dataset
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GET %s status %d", url, resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func (s *HFSync) uploadFile(path string, data []byte) error {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	fw, err := w.CreateFormFile("file", filepath.Base(path))
	if err != nil {
		return err
	}
	if _, err := fw.Write(data); err != nil {
		return err
	}
	if err := w.WriteField("pathInRepo", path); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	url := fmt.Sprintf("https://huggingface.co/api/datasets/%s/upload", s.repo)
	req, _ := http.NewRequest("POST", url, &buf)
	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("POST %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("POST %s status %d: %s", url, resp.StatusCode, string(body))
	}
	return nil
}

// --- Public API ---

// DownloadAll pulls the SQLite DB and assets from the HF Dataset on startup.
func (s *HFSync) DownloadAll() {
	log.Printf("[sync] downloading data from HF Dataset: %s", s.repo)

	if data, err := s.downloadFile("storage.db"); err != nil {
		log.Printf("[sync] error downloading storage.db: %v", err)
	} else if data != nil {
		if err := os.WriteFile(s.dbPath, data, 0644); err != nil {
			log.Printf("[sync] error writing storage.db: %v", err)
		} else {
			log.Println("[sync] restored storage.db from HF Dataset")
		}
	} else {
		log.Println("[sync] no existing storage.db in dataset, starting fresh")
	}
}

// SyncAll uploads the SQLite DB and any new/updated asset files.
func (s *HFSync) SyncAll() {
	// Upload DB if modified
	if data, err := os.ReadFile(s.dbPath); err == nil {
		info, _ := os.Stat(s.dbPath)
		modTime := info.ModTime()
		key := "storage.db"
		if lastMod, ok := s.lastSync[key]; !ok || modTime.After(lastMod) {
			if err := s.uploadFile(key, data); err != nil {
				log.Printf("[sync] upload %s error: %v", key, err)
			} else {
				s.lastSync[key] = modTime
				log.Println("[sync] uploaded storage.db")
			}
		}
	}

	// Upload new/modified asset files
	if err := os.MkdirAll(s.assetDir, 0755); err != nil {
		return
	}
	entries, _ := os.ReadDir(s.assetDir)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		key := "graph_assets/" + entry.Name()
		modTime := info.ModTime()

		if lastMod, ok := s.lastSync[key]; !ok || modTime.After(lastMod) {
			assetPath := filepath.Join(s.assetDir, entry.Name())
			if data, err := os.ReadFile(assetPath); err == nil {
				if err := s.uploadFile(key, data); err != nil {
					log.Printf("[sync] upload %s error: %v", key, err)
				} else {
					s.lastSync[key] = modTime
					log.Printf("[sync] uploaded %s", key)
				}
			}
		}
	}
}

// Start begins periodic background sync.
// DownloadAll should already have been called before opening the DB.
func (s *HFSync) Start(interval time.Duration) {
	if s == nil {
		return
	}
	go func() {
		s.SyncAll()

		ticker := time.NewTicker(interval)
		for range ticker.C {
			s.SyncAll()
		}
	}()
}
