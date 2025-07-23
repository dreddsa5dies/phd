package utils

import (
	"os"

	"github.com/dreddsa5dies/phd/Dissertation/code/src/models"
	"gopkg.in/yaml.v2"
)

type SimulationConfig struct {
	Steps         int     `yaml:"steps"`
	PoissonLambda float64 `yaml:"poisson_lambda"`
}

type Config struct {
	Machines   []models.Machine `yaml:"machines"`
	Tasks      []models.Task    `yaml:"tasks"`
	Simulation SimulationConfig `yaml:"simulation"`
}

// LoadConfig инициализирует конфигурацию из файла
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
