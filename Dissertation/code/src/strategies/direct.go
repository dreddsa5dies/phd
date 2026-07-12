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

// Стратегия заключается в последовательном проходе
// по машинам и задачам и их исполнения
func (s *DirectSearch) Run(machines []models.Machine, tasks []models.Task) StepMetrics {
	tNow := time.Now()

	// Инициализация метрик шага
	stepMetrics := StepMetrics{}
	machineStats := make(map[string]map[string]float64)

	// Флаг: смогла ли какая-то машина внести вклад на итерации
	var progress bool

	// Внутренний цикл итераций - пока есть прогресс
	for {
		progress = false

		// Каждая машина пытается внести вклад (в порядке списка)
		for i := range machines {
			// Пропускаем, если машина без энергии
			if machines[i].Energy <= 0 {
				continue
			}

			// Перебираем задачи в порядке следования
			for j := range tasks {
				// Пропускаем выполненную задачу
				if tasks[j].EnergyCost <= 0 {
					continue
				}

				// Проверка оборудования
				if !models.IntersectTypeEquipment(tasks[j].RequiredEquipment, machines[i].Equipment) {
					continue
				}

				// статистика по машина+задача
				taskStats := make(map[string]float64)

				// Машина вносит вклад: min(своей энергии, остатка задачи)
				contribution := machines[i].Energy
				if contribution > tasks[j].EnergyCost {
					contribution = tasks[j].EnergyCost
				}

				// Вносим вклад
				tasks[j].EnergyCost -= contribution
				machines[i].Energy -= contribution
				stepMetrics.EnergyUsed += contribution
				progress = true // был прогресс

				// Назначаем задачу как выполняемую (для отслеживания)
				machines[i].AssignedTask = &tasks[j]

				// Если задача выполнена - увеличиваем счетчик
				if tasks[j].EnergyCost <= 0 {
					stepMetrics.LenTasksDone++
					stepMetrics.TasksDone = append(stepMetrics.TasksDone, tasks[j].ID)
				}

				taskStats[tasks[j].ID] += contribution
				machineStats[machines[i].ID] = taskStats
				// После вклада - машина может продолжить в следующей итерации,
				// но на этой итерации она работает только с первой подходящей задачей
				// (по принципу "прямого перебора" - одна попытка за итерацию)
				break
			}
		}

		// Если ни одна машина не смогла внести вклад - выходим
		if !progress {
			break
		}
	}

	// Подсчет оставшихся невыполненных задач
	for _, task := range tasks {
		if task.EnergyCost > 0 {
			stepMetrics.LenNotExecTasks++
		}
	}

	stepMetrics.RealTime = time.Since(tNow).String()
	stepMetrics.MachineStats = machineStats

	// возврат метрик
	return stepMetrics
}
