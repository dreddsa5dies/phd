package strategies

import (
	"sort"
	"time"

	"github.com/dreddsa5dies/phd/Dissertation/code/src/models"
)

// MarketStrategy - реализация простого рынка / контракт-нет (inspired by Dias 2006)
type MarketStrategy struct{}

func (s *MarketStrategy) String() string {
	return "Рыночных методов"
}

// простой "cost" для ставки: чем меньше - лучше
// cost = (remainingTaskEnergy) / (availableMachineEnergy + eps)
// то есть машина с большим запасом энергии даёт более "дешёвую" ставку
func bidCost(taskEnergy float64, machineEnergy float64) float64 {
	eps := 1e-6
	return taskEnergy / (machineEnergy + eps)
}

// Run проводит один полный прогон стратегии; в отчёт записывает итоговые метрики
func (s *MarketStrategy) Run(machines []models.Machine, tasks []models.Task) StepMetrics {
	start := time.Now()

	// статистика по машинам+задачам
	machineStats := make(map[string]map[string]float64)

	// начальная сумма энергии
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

	doneSet := make(map[string]struct{})
	var doneList []string

	// Простая реализация аукционов: для каждой задачи запрашиваем ставки у всех подходящих машин.
	// Затем выбираем победителя(ей). Для крупной задачи можно распределить между несколькими машинами пропорционально их energy.
	// Проходим задачами детерминированно (по индексу) - это даёт воспроизводимость.
	for ti := range tasks {
		t := &tasks[ti]
		if t.EnergyCost <= 0 {
			continue
		}

		// собрать кандидатов
		type bid struct {
			machineIdx int
			cost       float64
		}
		bids := make([]bid, 0, len(machines))
		for mi := range machines {
			m := &machines[mi]
			if m.Energy <= 0 {
				continue
			}
			if !models.IntersectTypeEquipment(t.RequiredEquipment, m.Equipment) {
				continue
			}
			c := bidCost(t.EnergyCost, m.Energy)
			bids = append(bids, bid{machineIdx: mi, cost: c})
		}

		if len(bids) == 0 {
			// никто не может выполнить задачу
			continue
		}

		// сортируем по возрастанию стоимости (меньше cost - лучше)
		sort.Slice(bids, func(i, j int) bool {
			return bids[i].cost < bids[j].cost
		})

		// Если самая дешевая ставка способна завершить задачу - отдаем ей задачу полностью.
		// Иначе распределяем между топ-K машин пропорционально их энергии.
		top := bids[0]
		mTop := &machines[top.machineIdx]
		if mTop.Energy >= t.EnergyCost {
			// простое присвоение
			delta := t.EnergyCost
			if delta > mTop.Energy {
				delta = mTop.Energy
			}
			mTop.Energy -= delta
			t.EnergyCost -= delta
			if t.EnergyCost <= 0 {
				t.EnergyCost = 0
				if _, ok := doneSet[t.ID]; !ok {
					doneSet[t.ID] = struct{}{}
					doneList = append(doneList, t.ID)
				}
			}

			// фиксация статистики
			if machineStats[mTop.ID] == nil {
				machineStats[mTop.ID] = make(map[string]float64)
			}
			machineStats[mTop.ID][t.ID] += delta

			continue
		}

		// Иначе распределяем по нескольким машинам: берем топ-N (включая тех, у кого >0 energy)
		// Берём пока не покроем задачу или победителей не останется.
		remaining := t.EnergyCost
		// собираем список подходящих машин в порядке предпочтения
		preferred := make([]int, 0, len(bids))
		for _, b := range bids {
			preferred = append(preferred, b.machineIdx)
		}

		// простая схема: пропорционально энергии (bounded)
		totalAvail := 0.0
		for _, mi := range preferred {
			totalAvail += machines[mi].Energy
		}
		if totalAvail <= 0 {
			// нет доступной энергии
			continue
		}

		for _, mi := range preferred {
			if remaining <= 0 {
				break
			}
			m := &machines[mi]
			if m.Energy <= 0 {
				continue
			}
			// доля этой машины
			share := (m.Energy / totalAvail) * t.EnergyCost
			// ограничим по оставшейся энергии машины и по оставшейся задаче
			if share > m.Energy {
				share = m.Energy
			}
			if share > remaining {
				share = remaining
			}

			// фиксация статистики
			if machineStats[m.ID] == nil {
				machineStats[m.ID] = make(map[string]float64)
			}
			machineStats[m.ID][t.ID] += share

			m.Energy -= share
			remaining -= share
		}
		// уменьшение задачи на внесённый вклад
		done := (t.EnergyCost - remaining)
		if done > 0 {
			if remaining <= 0 {
				t.EnergyCost = 0
				if _, ok := doneSet[t.ID]; !ok {
					doneSet[t.ID] = struct{}{}
					doneList = append(doneList, t.ID)
				}
			} else {
				t.EnergyCost = remaining
			}
		}
	}

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

	return stepMetrics
}
