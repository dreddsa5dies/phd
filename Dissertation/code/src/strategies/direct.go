package strategies

import (
	"time"

	"github.com/dreddsa5dies/phd/Dissertation/code/src/models"
)

// стратегия прямого перебора
type DirectSearch struct{}

func (s *DirectSearch) String() string {
	return "Прямого перебора"
}

func (s *DirectSearch) Run(step int, machines []models.Machine, tasks []models.Task, report map[string]Metrics) {
	tNow := time.Now()

	mainMetrics, ok := report[s.String()]
	if !ok {
		return
	}

	var totalEnergyUsed float64
	var tasksDone int

	// Остаточная энергия для каждой задачи
	remainingEnergy := make(map[string]float64)
	for _, task := range tasks {
		remainingEnergy[task.ID] = task.EnergyCost
	}

	// Флаг: смогла ли какая-то машина внести вклад на итерации
	var progress bool

	// Внутренний цикл итераций — пока есть прогресс
	for {
		progress = false

		// Каждая машина пытается внести вклад (в порядке списка)
		for _, machine := range machines {
			// Пропускаем, если машина без энергии
			if machine.Energy <= 0 {
				continue
			}

			// Перебираем задачи в порядке следования
			for _, task := range tasks {
				// Пропускаем выполненную задачу
				if remainingEnergy[task.ID] <= 0 {
					continue
				}

				// Проверка оборудования
				if !models.IntersectTypeEquipment(task.RequiredEquipment, machine.Equipment) {
					continue
				}

				// Машина вносит вклад: min(своей энергии, остатка задачи)
				contribution := machine.Energy
				if contribution > remainingEnergy[task.ID] {
					contribution = remainingEnergy[task.ID]
				}

				// Вносим вклад
				remainingEnergy[task.ID] -= contribution
				machine.Energy -= contribution
				totalEnergyUsed += contribution
				progress = true // был прогресс

				// Назначаем задачу как выполняемую (для отслеживания)
				machine.AssignedTask = &task

				// Если задача выполнена — увеличиваем счётчик
				if remainingEnergy[task.ID] <= 0 {
					tasksDone++
				}

				// После вклада — машина может продолжить в следующей итерации,
				// но на этой итерации она работает только с первой подходящей задачей
				// (по принципу "прямого перебора" — одна попытка за итерацию)
				break
			}
		}

		// Если ни одна машина не смогла внести вклад — выходим
		if !progress {
			break
		}
	}

	// Подсчёт оставшихся невыполненных задач
	notExecTasks := 0
	for _, task := range tasks {
		if remainingEnergy[task.ID] > 0 {
			notExecTasks++
		}
	}

	// Сохранение метрик
	stepMetrics := StepMetrics{
		Step:         step,
		TasksDone:    tasksDone,
		EnergyUsed:   totalEnergyUsed,
		NotExecTasks: notExecTasks,
		Time:         time.Since(tNow).String(),
	}

	mainMetrics.StepMetric = append(mainMetrics.StepMetric, stepMetrics)
	report[s.String()] = mainMetrics
}
