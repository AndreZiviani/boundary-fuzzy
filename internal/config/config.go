// package config stores configuration around
package config

import (
	"encoding/json"
	"os"
	"path"

	"github.com/faabiosr/cachego/file"
)

type Config struct {
	AppName   string
	Favorites []string
}

func NewConfig() (Config, error) {
	c := Config{
		AppName: "boundary-fuzzy",
	}
	err := c.SetupConfigFolder()

	return c, err
}

// checks and or creates the config folder on startup
func (c Config) SetupConfigFolder() error {
	configFolder, err := c.ConfigFolder()
	if err != nil {
		return err
	}
	if _, err := os.Stat(configFolder); os.IsNotExist(err) {
		err := os.Mkdir(configFolder, 0700)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c Config) ConfigFolder() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	// check if the config folder already exists
	return path.Join(home, "."+c.AppName), nil
}

func (c *Config) LoadFavorites() error {
	configFolder, err := c.ConfigFolder()
	if err != nil {
		return err
	}

	cache := file.New(configFolder)
	favorites, err := cache.Fetch("favorites")
	if err != nil {
		// could not open file, we probably dont have favorites set up or it was removed, ignoring
		return nil
	}

	if err := json.Unmarshal([]byte(favorites), &c.Favorites); err != nil {
		return err
	}

	return err
}

func (c *Config) SaveFavorites() error {
	configFolder, err := c.ConfigFolder()
	if err != nil {
		return err
	}

	cache := file.New(configFolder)
	favorites, err := json.Marshal(c.Favorites)
	if err != nil {
		return err
	}

	if err := cache.Save("favorites", string(favorites), 0); err != nil {
		return err
	}

	return nil
}
