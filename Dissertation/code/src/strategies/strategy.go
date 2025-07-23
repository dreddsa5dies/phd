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
	// Общее энергоемкость задач
	TotalEnergyTasks float64 `json:"totalEnergyTasks"`
	// Общее время выполнения
	TotalTime string `json:"totalTime"`
	// метрики каждого шага
	StepMetric []StepMetrics `json:"stepMetrics"`
}

type StepMetrics struct {
	// Номер прогона
	Step int `json:"step"`
	// Количество выполненных задач
	TasksDone int `json:"tasksDone"`
	// Энергозатраты машин
	EnergyUsed float64 `json:"energyUsed"`
	// Не исполненные задачи
	NotExecTasks int `json:"notExecTasks"`
	// Время выполнения
	Time string `json:"time"`
}

// общий интерфейс стратегий по выбору задач
type Strategy interface {
	Run(step int, machines []models.Machine, tasks []models.Task, report map[string]Metrics)
	String() string
}
