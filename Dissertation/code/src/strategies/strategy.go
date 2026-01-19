package strategies

import (
	"github.com/dreddsa5dies/phd/Dissertation/code/src/models"
)

// Метрики
type Metrics struct {
	// Общее количество задач всего
	TotalTasks int `json:"totalTasks"`
	// Общее количество машин всего
	TotalMachine int `json:"totalMachine"`
	// Общая энергоемкость машин
	TotalEnergyMachines float64 `json:"totalEnergyMachines"`
	// Список машин
	AllMachines []models.Machine `json:"allMachines"`
	// Общее энергоемкость задач
	TotalEnergyTasks float64 `json:"totalEnergyTasks"`
	// Список задач
	AllTasks []models.Task `json:"allTasks"`
	// Общее время выполнения
	TotalTime string `json:"totalTime"`
	// Метрики каждого шага
	StepMetric []StepMetrics `json:"stepMetrics"`
}

type StepMetrics struct {
	Step int `json:"step"` // Номер прогона
	// Количество выполненных задач
	LenTasksDone int `json:"lenTasksDone"`
	// Перечень исполненных задач по ID
	TasksDone []string `json:"tasksDone"`
	// Энергозатраты машин
	EnergyUsed float64 `json:"energyUsed"`
	// Не исполненные задачи
	LenNotExecTasks int `json:"lenNotExecTasks"`
	// Реальное время выполнения шага
	RealTime string `json:"realTime"`
	// детализация по машинам:
	// string1 -> machineID, string2 -> taskID, float64 -> потрачено
	MachineStats map[string]map[string]float64 `json:"machineStats"`
}

// общий интерфейс стратегий по выбору задач
type Strategy interface {
	Run(machines []models.Machine, tasks []models.Task) StepMetrics
	String() string
}
