package strategies

import (
	"math"
	"time"

	"github.com/dreddsa5dies/phd/Dissertation/code/src/models"
)

// DelayMinimizationStrategy реализует стратегию минимизации временных задержек
type DelayMinimizationStrategy struct {
	Alpha float64 // вес координации
	Beta  float64 // вес: задержка * сложность
}

func (s *DelayMinimizationStrategy) String() string {
	return "Минимизация ВрЗ"
}

func (s *DelayMinimizationStrategy) Run(machines []models.Machine, tasks []models.Task) StepMetrics {
	start := time.Now()

	// Соберём начальную энергию (до прогонки)
	sumEnergy := func() float64 {
		var sum float64
		for i := range machines {
			if machines[i].Energy > 0 {
				sum += machines[i].Energy
			}
		}
		return sum
	}
	initialTotalEnergy := sumEnergy()

	// Вспом. функции
	countUnfinished := func() int {
		var n int
		for i := range tasks {
			if tasks[i].EnergyCost > 0 {
				n++
			}
		}
		return n
	}
	computeW := func(t *models.Task) float64 {
		age := time.Now().Unix() - t.CreatedAt
		if age < 0 {
			age = 0
		}
		return float64(age)
	}

	totalMachines := len(machines)
	maxIters := 10_000 + 10*len(tasks)*maximum(1, totalMachines)

	// prevChoices для оценки Pj (частотная оценка предыдущей итерации)
	prevChoices := make(map[string]int)

	// Набор завершённых задач за весь прогон
	doneSet := make(map[string]struct{})
	var doneList []string

	// Внутренний итерационный цикл стратегии — выполняем до сходимости,
	// но в отчёт запишем только итог за весь прогон.
	for iter := 0; iter < maxIters; iter++ {
		// Стоп-критерии
		if countUnfinished() == 0 || sumEnergy() == 0 {
			break
		}

		// --- ФАЗА ВЫБОРА (последовательно, детерминированно) ---
		// Для каждой машины выбираем индекс задачи или -1
		choices := make([]int, len(machines))
		for i := range choices {
			choices[i] = -1
		}

		for i := range machines {
			m := &machines[i]
			if m.Energy <= 0 {
				continue
			}

			bestIdx := -1
			bestVal := math.Inf(1)

			for j := range tasks {
				t := &tasks[j]
				if t.EnergyCost <= 0 {
					continue
				}
				if !models.IntersectTypeEquipment(t.RequiredEquipment, m.Equipment) {
					continue
				}
				Wj := computeW(t)
				Qj := t.EnergyCost
				var Pj float64
				if totalMachines > 0 {
					Pj = float64(prevChoices[t.ID]) / float64(totalMachines)
				}
				val := s.Beta*Wj*Qj - s.Alpha*Pj
				if val < bestVal {
					bestVal = val
					bestIdx = j
				}
			}
			choices[i] = bestIdx
		}

		// --- Агрегация выборов ---
		// taskIdx -> []machineIdx
		chosenMap := make(map[int][]int)
		currChoices := make(map[string]int)
		for mi, tIdx := range choices {
			if tIdx >= 0 {
				chosenMap[tIdx] = append(chosenMap[tIdx], mi)
				currChoices[tasks[tIdx].ID]++
			}
		}

		// --- ФАЗА ИСПОЛНЕНИЯ: только машины, выбравшие задачу, вносят вклад ---
		var energySpentThisIter float64
		for tIdx := range tasks {
			t := &tasks[tIdx]
			if t.EnergyCost <= 0 {
				continue
			}
			mList := chosenMap[tIdx]
			if len(mList) == 0 {
				continue
			}

			initialCost := t.EnergyCost
			var totalContrib float64

			// Машины последовательным образом вносят вклад (чтобы избежать гонок)
			for _, mi := range mList {
				if mi < 0 || mi >= len(machines) {
					continue
				}
				m := &machines[mi]
				if m.Energy <= 0 {
					continue
				}

				delta := m.Energy
				remaining := initialCost - totalContrib
				if delta > remaining {
					delta = remaining
				}
				// уменьшаем энергию машины и добавляем вклад
				m.Energy -= delta
				totalContrib += delta
				if totalContrib >= initialCost {
					break
				}
			}

			energySpentThisIter += totalContrib

			if totalContrib >= initialCost {
				// задача завершена
				t.EnergyCost = 0
				if _, ok := doneSet[t.ID]; !ok {
					doneSet[t.ID] = struct{}{}
					doneList = append(doneList, t.ID)
				}
			} else if totalContrib > 0 {
				// частично выполнена
				t.EnergyCost = initialCost - totalContrib
			}
		}

		// Обновляем prevChoices для следующей итерации
		prevChoices = make(map[string]int)
		for k, v := range currChoices {
			prevChoices[k] = v
		}

		// Если прогресса нет — завершаем
		if energySpentThisIter == 0 {
			break
		}
	}

	// Подсчёт итоговых метрик за весь прогон (один StepMetrics)
	finalTotalEnergy := sumEnergy()
	energyUsed := initialTotalEnergy - finalTotalEnergy

	stepMetrics := StepMetrics{
		EnergyUsed:      energyUsed,
		TasksDone:       doneList,
		LenTasksDone:    len(doneList),
		LenNotExecTasks: 0, // заполним ниже
		RealTime:        time.Since(start).String(),
	}

	for i := range tasks {
		if tasks[i].EnergyCost > 0 {
			stepMetrics.LenNotExecTasks++
		}
	}

	return stepMetrics
}

// maximum — вспомогательная функция вычисления максимума
func maximum(a, b int) int {
	if a >= b {
		return a
	}
	return b
}
