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
		AllMachines:         machines,
		TotalEnergyTasks:    models.TotalEnergyTasks(tasks),
		AllTasks:            tasks,
		TotalTime:           "",
	}

	// прогон стратегии заданное количество шагов
	runMetrics := make([]strategies.StepMetrics, steps)
	for i := range steps {
		// Копируем список задач и машин
		// независимые прогоны стратегий
		taskList := make([]models.Task, len(tasks))
		copy(taskList, tasks)

		machineList := make([]models.Machine, len(machines))
		copy(machineList, machines)

		runMetrics[i] = strategy.Run(machineList, taskList)
		runMetrics[i].Step = i + 1
	}

	metrics.StepMetric = runMetrics
	// фиксация итогового времени
	metrics.TotalTime = time.Since(tNow).String()

	report[strategy.String()] = metrics
}
