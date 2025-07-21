package utils

import (
	"os"

	"gopkg.in/yaml.v2"
)

type RobotConfig struct {
	ID        string  `yaml:"id"`
	Equipment string  `yaml:"equipment"`
	Energy    float64 `yaml:"energy"`
}

type TaskConfig struct {
	ID                string  `yaml:"id"`
	RequiredEquipment string  `yaml:"required_equipment"`
	EnergyCost        float64 `yaml:"energy_cost"`
	Priority          int     `yaml:"priority"`
	Deadline          int     `yaml:"deadline"`
}

type SimulationConfig struct {
	MaxSteps      int     `yaml:"max_steps"`
	Strategy      string  `yaml:"strategy"`
	PoissonLambda float64 `yaml:"poisson_lambda"`
}

type Config struct {
	Robots     []RobotConfig    `yaml:"robots"`
	Tasks      []TaskConfig     `yaml:"tasks"`
	Simulation SimulationConfig `yaml:"simulation"`
}

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
