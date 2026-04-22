package cookies

import (
	"embed"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Laky-64/gologging"
	"resty.dev/v3"

	"main/internal/config"
)

const cookieDir = "internal/cookies"

var (
	cachedFiles []string
	cacheOnce   sync.Once
	client      = resty.New()
)

//go:embed *.txt
var embeddedCookies embed.FS

func init() {
	gologging.Debug("🔹 Initializing cookies...")

	if err := copyEmbeddedCookies(); err != nil {
		gologging.Fatal("Failed to copy embedded cookies:", err)
	}

	urls := strings.Fields(config.CookiesLink)
	for _, url := range urls {
		if err := downloadCookieFile(url); err != nil {
			gologging.WarnF(
				"Failed to download cookie file from %s: %v",
				url,
				err,
			)
		}
	}
}

func copyEmbeddedCookies() error {
	entries, err := embeddedCookies.ReadDir(".")
	if err != nil {
		return err
	}

	for _, e := range entries {

		if e.IsDir() || e.Name() == "example.txt" {
			continue
		}

		dst := filepath.Join(cookieDir, e.Name())

		if _, err := os.Stat(dst); err == nil {
			continue
		}

		data, err := embeddedCookies.ReadFile(e.Name())
		if err != nil {
			return err
		}

		if err := os.WriteFile(dst, data, 0o600); err != nil {
			return err
		}
	}

	return nil
}

func downloadCookieFile(url string) error {
	id := filepath.Base(url)
	rawURL := "https://batbin.me/raw/" + id
	filePath := filepath.Join(cookieDir, id+".txt")

	resp, err := client.R().
		SetOutputFileName(filePath).
		Get(rawURL)
	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf(
			"unexpected status %d from %s",
			resp.StatusCode(),
			rawURL,
		)
	}

	return nil
}

func loadCookieCache() error {
	files, err := filepath.Glob(filepath.Join(cookieDir, "*.txt"))
	if err != nil {
		return err
	}

	var filtered []string

	for _, f := range files {
		if filepath.Base(f) == "example.txt" {
			continue
		}
		filtered = append(filtered, f)
	}

	cachedFiles = filtered
	return nil
}

func GetRandomCookieFile() (string, error) {
	var err error

	cacheOnce.Do(func() {
		err = loadCookieCache()
	})

	if err != nil {
		gologging.WarnF("Failed to load cookie cache: %v", err)
		return "", err
	}

	if len(cachedFiles) == 0 {
		gologging.Warn("No cookie files available")
		return "", nil
	}

	return cachedFiles[rand.Intn(len(cachedFiles))], nil
}
