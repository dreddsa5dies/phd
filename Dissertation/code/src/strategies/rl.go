package strategies

import (
	"time"

	"github.com/dreddsa5dies/phd/Dissertation/code/src/models"
)

// стратегия по алгоритмам на основе RL (Кошманова, 2012)
type RLStrategy struct {
	Rewards map[string]float64
}

func (s *RLStrategy) String() string {
	return "RL"
}

// TODO переделать
func (s *RLStrategy) Run(step int, machines []models.Machine, tasks []models.Task, report map[string]Metrics) {
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
