package strategies

import (
	"time"

	"github.com/dreddsa5dies/phd/Dissertation/code/src/models"
)

// стратегия по рыночным методам (Dias, 2006)
type MarketStrategy struct{}

func (s *MarketStrategy) String() string {
	return "Рыночных методов"
}

// TODO переделать
func (s *MarketStrategy) Run(step int, machines []models.Machine, tasks []models.Task, report map[string]Metrics) {
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
