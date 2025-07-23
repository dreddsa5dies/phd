package strategies

import (
	"time"

	"github.com/dreddsa5dies/phd/Dissertation/code/src/models"
)

// энергосберегающая стратегия
type EnergyEfficientStrategy struct{}

func (s *EnergyEfficientStrategy) String() string {
	return "Энергосберегающая"
}

// TODO переделать
func (s *EnergyEfficientStrategy) Run(step int, machines []models.Machine, tasks []models.Task, report map[string]Metrics) {
	tNow := time.Now()

	mainMetrics, ok := report[s.String()]
	if !ok {
		return
	}

	stepMetrics := StepMetrics{
		Step:         step,
		TasksDone:    0,
		EnergyUsed:   0,
		NotExecTasks: 0,
		Time:         time.Since(tNow).String(),
	}

	// сохранение результата
	mainMetrics.StepMetric = append(mainMetrics.StepMetric, stepMetrics)

	report[s.String()] = mainMetrics
}
