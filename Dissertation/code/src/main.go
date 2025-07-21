package main

import (
	"fmt"
	"log"

	"github.com/dreddsa5dies/phd/Dissertation/code/src/engine"
	"github.com/dreddsa5dies/phd/Dissertation/code/src/strategies"
	"github.com/dreddsa5dies/phd/Dissertation/code/src/utils"
)

func main() {
	config, err := utils.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	var results []engine.SimulationResult

	strategies := []struct {
		name     string
		strategy strategies.Strategy
	}{
		{"Energy Efficient", &strategies.EnergyEfficientStrategy{}},
		{"Delay Minimization", &strategies.DelayMinimizationStrategy{}},
		{"Market-Based", &strategies.MarketStrategy{}},
		{"RL-Based", &strategies.RLStrategy{Rewards: make(map[string]float64)}},
	}

	for _, s := range strategies {
		fmt.Printf("Running simulation with strategy: %s\n", s.name)
		result := engine.RunSimulation(config, s.strategy)
		result.StrategyName = s.name
		results = append(results, result)
	}

	engine.GenerateReport(results)
	fmt.Println("Report generated: report.csv")
}
