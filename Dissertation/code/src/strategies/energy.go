package strategies

import (
	"math"
	"time"

	"github.com/dreddsa5dies/phd/Dissertation/code/src/models"
)

// энергосберегающая стратегия
type EnergyEfficientStrategy struct{}

// String возвращает имя стратегии
func (s *EnergyEfficientStrategy) String() string {
	return "Энергосберегающая"
}

func (s *EnergyEfficientStrategy) Run(machines []models.Machine, tasks []models.Task) StepMetrics {
	start := time.Now()

	// статистика по машинам
	machineStats := make(map[string]map[string]float64)

	// Подсчёт начальной энергии
	sumEnergy := func() float64 {
		var sum float64
		for i := range machines {
			if machines[i].Energy > 0 {
				sum += machines[i].Energy
			}
		}
		return sum
	}
	initialEnergy := sumEnergy()

	// Набор выполненных задач
	doneSet := make(map[string]struct{})
	var doneList []string

	// Внутренний цикл выполнения до завершения или исчерпания энергии
	maxIters := 10_000 + 10*len(tasks)*maximum(1, len(machines))
	for iter := 0; iter < maxIters; iter++ {
		if sumEnergy() == 0 {
			break
		}

		progress := false

		// ФАЗА ВЫБОРА + ИСПОЛНЕНИЯ
		for i := range machines {
			m := &machines[i]
			if m.Energy <= 0 {
				continue
			}

			// выбор доступной задачи с минимальным Qj/Ri
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

				val := t.EnergyCost / m.Energy
				if val < bestVal {
					bestVal = val
					bestIdx = j
				}
			}

			if bestIdx == -1 {
				continue
			}

			// Исполнение выбранной задачи
			t := &tasks[bestIdx]
			initialCost := t.EnergyCost
			delta := m.Energy
			if delta > initialCost {
				delta = initialCost
			}

			// фиксация статистики по машина+задача
			if machineStats[m.ID] == nil {
				machineStats[m.ID] = make(map[string]float64)
			}
			machineStats[m.ID][t.ID] += delta

			m.Energy -= delta
			t.EnergyCost -= delta
			progress = true

			if t.EnergyCost <= 0 {
				t.EnergyCost = 0
				if _, ok := doneSet[t.ID]; !ok {
					doneSet[t.ID] = struct{}{}
					doneList = append(doneList, t.ID)
				}
			}
		}

		if !progress {
			break
		}
	}

	// Итоговые метрики
	finalEnergy := sumEnergy()
	energyUsed := initialEnergy - finalEnergy

	stepMetrics := StepMetrics{
		EnergyUsed:      energyUsed,
		TasksDone:       doneList,
		LenTasksDone:    len(doneList),
		LenNotExecTasks: 0,
		RealTime:        time.Since(start).String(),
		MachineStats:    machineStats, // сохранение статистики
	}

	for i := range tasks {
		if tasks[i].EnergyCost > 0 {
			stepMetrics.LenNotExecTasks++
		}
	}

	// возврат метрик
	return stepMetrics
}
