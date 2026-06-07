package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type GoxConfig struct {
	Name    string          `mapstructure:"name"`
	Apps    map[string]App  `mapstructure:"apps"`
	Libs    map[string]Lib    `mapstructure:"libs"`
	Tasks    map[string][]Task `mapstructure:"tasks"`
	Release  Release           `mapstructure:"release"`
	Profiles map[string]Profile`mapstructure:"profiles"`
}

type Profile struct {
	Ldflags []string          `mapstructure:"ldflags"`
	Env     map[string]string `mapstructure:"env"`
	Race    bool              `mapstructure:"race"`
}

type App struct {
	Path string `mapstructure:"path"`
	Main string `mapstructure:"main"`
	Port string `mapstructure:"port"`
	Env  string `mapstructure:"env"`
}

type Lib struct {
	Path string `mapstructure:"path"`
}

type Task struct {
	Run string `mapstructure:"run"`
}

type Release struct {
	Targets []string `mapstructure:"targets"`
}

func LoadConfig(configPath string) (*GoxConfig, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("Can not read config: %w", err)
	}

	var cfg GoxConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("Can not parse config into struct: %w", err)
	}

	return &cfg, nil
}
