package engine

import (
	"time"

	"github.com/dreddsa5dies/phd/Dissertation/code/src/models"
	"github.com/dreddsa5dies/phd/Dissertation/code/src/strategies"
)

// Прогон экспериментов
func RunSimulation(machines []models.Machine, tasks []models.Task, strategy strategies.Strategy, steps int, report map[string]strategies.Metrics) {
	// подготовка отчета
	tNow := time.Now()
	metrics := strategies.Metrics{
		TotalTasks:          len(tasks),
		TotalMachine:        len(machines),
		TotalEnergyMachines: models.TotalEnergyMachines(machines),
		TotalEnergyTasks:    models.TotalEnergyTasks(tasks),
		TotalTime:           "",
	}

	report[strategy.String()] = metrics

	// прогон стратегии заданное количество шагов
	for i := range steps {
		strategy.Run(i, machines, tasks, report)
	}

	// фиксация итогового времени
	v := report[strategy.String()]
	v.TotalTime = time.Since(tNow).String()
	report[strategy.String()] = v
}
