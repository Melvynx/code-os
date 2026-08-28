package screenshots

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/melvynx/code-os/internal/model"
)

var supported = map[string]bool{
	".avif": true, ".gif": true, ".jpeg": true, ".jpg": true,
	".png": true, ".svg": true, ".webp": true,
}

type Indexer struct {
	Root string
}

func (indexer Indexer) Scan() ([]model.Screenshot, error) {
	root, err := filepath.Abs(indexer.Root)
	if err != nil {
		return nil, fmt.Errorf("resolve screenshot root: %w", err)
	}
	if _, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			return []model.Screenshot{}, nil
		}
		return nil, fmt.Errorf("read screenshot root: %w", err)
	}
	var images []model.Screenshot
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !supported[strings.ToLower(filepath.Ext(entry.Name()))] {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		parts := strings.Split(relative, string(filepath.Separator))
		group := "screenshots"
		project := ""
		if len(parts) > 1 {
			group = parts[0]
		}
		if len(parts) > 2 {
			project = parts[1]
		}
		id := imageID(path)
		images = append(images, model.Screenshot{
			ID: id, Name: entry.Name(), Path: path, URL: "/media/" + url.PathEscape(id),
			Project: project, Group: group, Size: info.Size(), CreatedAt: info.ModTime(),
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("index screenshots: %w", err)
	}
	sort.Slice(images, func(i, j int) bool { return images[i].CreatedAt.After(images[j].CreatedAt) })
	return images, nil
}

func imageID(path string) string {
	sum := sha256.Sum256([]byte(path))
	return hex.EncodeToString(sum[:10])
}
