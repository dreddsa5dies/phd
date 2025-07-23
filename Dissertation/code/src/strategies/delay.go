package strategies

import (
	"math"
	"time"

	"github.com/dreddsa5dies/phd/Dissertation/code/src/models"
)

// стратегия минимизации временных задержек
type DelayMinimizationStrategy struct {
	// параметры стратегии
	Alpha float64
	Beta  float64
}

func (s *DelayMinimizationStrategy) String() string {
	return "Минимизации временных задержек"
}

func (s *DelayMinimizationStrategy) Run(step int, machines []models.Machine, tasks []models.Task, report map[string]Metrics) {
	tNow := time.Now()

	mainMetrics, ok := report[s.String()]
	if !ok {
		mainMetrics = Metrics{StepMetric: []StepMetrics{}}
	}

	// Остаточная энергия для каждой задачи
	remainingEnergy := make(map[string]float64)
	for _, t := range tasks {
		remainingEnergy[t.ID] = t.EnergyCost
	}

	// Время ожидания (в условных единицах)
	waitTime := make(map[string]float64)
	for _, t := range tasks {
		waitTime[t.ID] = 0
	}

	// История выборов (для учёта "популярности" задачи)
	history := make(map[string]int)

	var tasksDoneTotal int
	var energyUsedTotal float64

	// Внутренние итерации — пока есть прогресс
	for {
		sharedMem := make(map[string][]models.Machine) // кто выбрал какую задачу

		// Каждая машина выбирает задачу
		for _, m := range machines {
			m.AssignedTask = nil

			// Формируем список кандидатов
			var candidates []models.Task
			for _, t := range tasks {
				if remainingEnergy[t.ID] <= 0 {
					continue
				}
				if models.IntersectTypeEquipment(t.RequiredEquipment, m.Equipment) && m.Energy > 0 {
					candidates = append(candidates, t)
				}
			}

			if len(candidates) == 0 {
				continue
			}

			// Вычисляем вероятность выбора на основе истории
			prob := make(map[string]float64)
			totalHistory := 0
			for _, t := range candidates {
				totalHistory += history[t.ID]
			}
			for _, t := range candidates {
				if totalHistory > 0 {
					prob[t.ID] = float64(history[t.ID]) / float64(totalHistory)
				} else {
					prob[t.ID] = 1.0 / float64(len(candidates))
				}
			}

			// Выбор задачи с минимальной стоимостью
			minCost := math.Inf(1)
			for _, t := range candidates {
				cost := s.Beta*waitTime[t.ID]*t.EnergyCost - s.Alpha*prob[t.ID]
				if cost < minCost {
					minCost = cost

					sharedMem[t.ID] = append(sharedMem[t.ID], m)
					m.AssignedTask = &t
				}
			}
		}

		// Обработка вкладов
		localProgress := false
		for _, t := range tasks {
			if remainingEnergy[t.ID] <= 0 {
				continue
			}

			machinesSelected := sharedMem[t.ID]
			if len(machinesSelected) == 0 {
				// Никто не выбрал — увеличиваем задержку
				waitTime[t.ID] += 1
				continue
			}

			// Каждая выбранная машина вносит вклад
			for _, m := range machinesSelected {
				if m.Energy <= 0 || remainingEnergy[t.ID] <= 0 {
					continue
				}

				contribution := m.Energy
				if contribution > remainingEnergy[t.ID] {
					contribution = remainingEnergy[t.ID]
				}

				remainingEnergy[t.ID] -= contribution
				m.Energy -= contribution
				energyUsedTotal += contribution
				localProgress = true

				// Одна попытка за итерацию
				break
			}

			// Проверка выполнения задачи
			if remainingEnergy[t.ID] <= 0 {
				tasksDoneTotal++
				history[t.ID]++ // увеличиваем популярность
			} else {
				// Задача не завершена — растёт задержка
				waitTime[t.ID] += 1
			}
		}

		// Если на этой итерации никто не внес вклад — выходим
		if !localProgress {
			break
		}
	}

	// Подсчёт оставшихся невыполненных задач
	notExecTasks := 0
	for _, t := range tasks {
		if remainingEnergy[t.ID] > 0 {
			notExecTasks++
		}
	}

	// Сохранение метрик
	stepMetrics := StepMetrics{
		Step:         step,
		TasksDone:    tasksDoneTotal,
		EnergyUsed:   energyUsedTotal,
		NotExecTasks: notExecTasks,
		Time:         time.Since(tNow).String(),
	}

	mainMetrics.StepMetric = append(mainMetrics.StepMetric, stepMetrics)
	report[s.String()] = mainMetrics
}
