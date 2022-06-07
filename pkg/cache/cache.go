package cache

import (
	"errors"
	"fmt"
	"github.com/ghodss/yaml"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Cache struct {
	RootDir string
}

var replacer = strings.NewReplacer("/", "_", ".", "_")

func (c *Cache) GetResourcePath(key string) string {
	maybeResourcePath := filepath.Join(c.RootDir, replacer.Replace(key))
	return maybeResourcePath
}

func (c *Cache) Exists(key string) (bool, error) {
	_, err := os.Stat(c.GetResourcePath(key))
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *Cache) Store(key string, data []byte) error {
	if _, err := os.Stat(c.RootDir); errors.Is(err, os.ErrNotExist) {
		mkdirErr := os.MkdirAll(c.RootDir, os.ModePerm)
		if mkdirErr != nil {
			return fmt.Errorf("error creating cache root dir: %w", mkdirErr)
		}
	}
	err := ioutil.WriteFile(c.GetResourcePath(key), data, 0644)
	if err != nil {
		return fmt.Errorf("error writing resource cache: %w", err)
	}
	return nil
}

func (c *Cache) GetCachedResource(key string, data []byte) ([]byte, error) {
	exists, err := c.Exists(key)
	if err != nil {
		return nil, fmt.Errorf("error checking if cache key exists: %w", err)
	}
	if exists {
		cachedBytes, err := ioutil.ReadFile(c.GetResourcePath(key))
		if err != nil {
			return nil, fmt.Errorf("error reading cached resource: %w", err)
		}
		return cachedBytes, nil
	}

	var out map[string]interface{}
	unmarshalErr := yaml.Unmarshal(data, &out)
	if unmarshalErr != nil {
		return nil, fmt.Errorf("error unmarshalling resource bytes: %w", unmarshalErr)
	}
	storeErr := c.Store(key, data)
	if storeErr != nil {
		return nil, storeErr
	}
	return data, nil
}
