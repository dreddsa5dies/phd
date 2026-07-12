package strategies

import (
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/dreddsa5dies/phd/Dissertation/code/src/models"
)

// RLStrategy - стратегия на базе Q-learning (модифицированная для задач распределения)
type RLStrategy struct {
	// Q: map[machineID]map[equipmentType]Qvalue
	Q map[string]map[int]float64

	// гиперпараметры обучения
	LR      float64 // learning rate
	Gamma   float64 // discount factor
	Epsilon float64 // начальное epsilon для epsilon-greedy
	MinEps  float64 // минимальное epsilon
	Decay   float64 // decay per Run call

	mu sync.Mutex // защита Q если стратегия используется параллельно
}

func (s *RLStrategy) String() string {
	return "RL (Q-learning)"
}

// Run выполняет один шаг (целый прогон стратегии).
// Реализует интерактивный Q-learning: машины действуют, получают награды, обновляют Q.
func (s *RLStrategy) Run(machines []models.Machine, tasks []models.Task) StepMetrics {
	start := time.Now()
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// статистика по машинам+задачам
	machineStats := make(map[string]map[string]float64)

	// вспомогательные функции
	sumEnergy := func() float64 {
		var sum float64
		for i := range machines {
			if machines[i].Energy > 0 {
				sum += machines[i].Energy
			}
		}
		return sum
	}
	countUnfinished := func() int {
		var n int
		for i := range tasks {
			if tasks[i].EnergyCost > 0 {
				n++
			}
		}
		return n
	}

	// Подготовка итоговых метрик
	doneSet := make(map[string]struct{})
	var doneList []string

	initialTotalEnergy := sumEnergy()

	// Параметры итераций внутри прогона
	maxIters := 10000 + 10*len(tasks)*maximum(1, len(machines))

	// Основной итерационный цикл: машины действуют и учатся
	for iter := 0; iter < maxIters; iter++ {
		// стоп-критерии
		if sumEnergy() == 0 || countUnfinished() == 0 {
			break
		}

		anyProgress := false

		// для каждой машины - выбор задачи и исполнение
		for mi := range machines {
			m := &machines[mi]
			if m.Energy <= 0 {
				continue
			}

			// Соберем доступные действия (equipment types задач, которые машина может делать)
			availableActions := make([]int, 0, 8)
			taskIdxByAction := make(map[int][]int) // action -> список индексов задач этого типа
			for ti := range tasks {
				t := &tasks[ti]
				if t.EnergyCost <= 0 {
					continue
				}
				if !models.IntersectTypeEquipment(t.RequiredEquipment, m.Equipment) {
					continue
				}
				act := int(t.RequiredEquipment)
				taskIdxByAction[act] = append(taskIdxByAction[act], ti)
			}
			for a := range taskIdxByAction {
				availableActions = append(availableActions, a)
			}
			if len(availableActions) == 0 {
				continue
			}

			// Инициализация Q для машины
			s.mu.Lock()
			s.ensureQ(m.ID)
			s.mu.Unlock()

			// epsilon-greedy выбор действия (equipment)
			var chosenAction int
			if rng.Float64() < s.Epsilon {
				// случайный выбор
				chosenAction = availableActions[rng.Intn(len(availableActions))]
			} else {
				// выбрать по Q
				s.mu.Lock()
				ca, _ := s.argMaxQ(m.ID, availableActions)
				s.mu.Unlock()
				chosenAction = ca
			}

			// в действии: выбрать конкретную задачу этого типа, например самую "дешевую" (по energy cost)
			candidates := taskIdxByAction[chosenAction]
			if len(candidates) == 0 {
				continue
			}
			// выбрать задачу с наименьшей energy_cost (энергосбережение мотивационно)
			bestTi := candidates[0]
			minCost := tasks[bestTi].EnergyCost
			for _, ti := range candidates {
				if tasks[ti].EnergyCost < minCost {
					minCost = tasks[ti].EnergyCost
					bestTi = ti
				}
			}

			// Выполнение: машина вносит всю свою энергию (или столько, сколько нужно)
			task := &tasks[bestTi]
			if m.Energy <= 0 || task.EnergyCost <= 0 {
				continue
			}
			delta := m.Energy
			if delta > task.EnergyCost {
				delta = task.EnergyCost
			}
			// расходы и вклад делаем в основном потоке - безопасно
			m.Energy -= delta
			task.EnergyCost -= delta
			anyProgress = anyProgress || delta > 0

			// сохранение статистики
			if machineStats[m.ID] == nil {
				machineStats[m.ID] = make(map[string]float64)
			}
			machineStats[m.ID][task.ID] += delta

			// Награда: положительная за завершение задачи; отрицательная за расход энергии
			var reward float64
			if task.EnergyCost <= 0 {
				// задача завершена
				task.EnergyCost = 0
				_, existed := doneSet[task.ID]
				if !existed {
					doneSet[task.ID] = struct{}{}
					doneList = append(doneList, task.ID)
				}
				// премия за завершение; масштабируем по размеру задачи
				reward = 10.0 + 0.01*float64(delta)
			} else {
				// частичное выполнение дает слабую награду
				reward = 0.01 * float64(delta)
			}
			// штраф за расход энергии (чтобы стимулировать экономию)
			reward -= 0.001 * float64(delta)

			// Q-update: Q(s,a) <- Q + lr*(reward + gamma * max_a' Q(s',a') - Q)
			// В данной простой модели состояние - остаток энергии машины (дискретизируем в бакеты)
			// Но для простоты используем одношаговый TD: следующая Q max для доступных действий
			var maxNext float64
			s.mu.Lock()
			// ensure entries exist for actions
			for _, a := range availableActions {
				if _, ok := s.Q[m.ID][a]; !ok {
					s.Q[m.ID][a] = 0.0
				}
				if s.Q[m.ID][a] > maxNext {
					maxNext = s.Q[m.ID][a]
				}
			}
			oldQ := s.Q[m.ID][chosenAction]
			newQ := oldQ + s.LR*(reward+s.Gamma*maxNext-oldQ)
			s.Q[m.ID][chosenAction] = newQ
			s.mu.Unlock()
		} // конец по машинам

		// уменьшение epsilon (мягкое)
		s.mu.Lock()
		if s.Epsilon > s.MinEps {
			s.Epsilon *= s.Decay
			if s.Epsilon < s.MinEps {
				s.Epsilon = s.MinEps
			}
		}
		s.mu.Unlock()

		if !anyProgress {
			break
		}
	} // конец итераций

	// итоговые метрики
	finalTotalEnergy := sumEnergy()
	energyUsed := initialTotalEnergy - finalTotalEnergy

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

	// безопасная запись в report
	return stepMetrics
}

// helper: ensure Q entry exists
func (s *RLStrategy) ensureQ(machineID string) {
	if _, ok := s.Q[machineID]; !ok {
		s.Q[machineID] = make(map[int]float64)
	}
}

// argmax over equipment actions for given machine state (available actions)
func (s *RLStrategy) argMaxQ(machineID string, available []int) (bestAction int, bestVal float64) {
	s.ensureQ(machineID)
	bestVal = math.Inf(-1)
	bestAction = -1
	for _, a := range available {
		v := s.Q[machineID][a]
		if v > bestVal || bestAction == -1 {
			bestVal = v
			bestAction = a
		}
	}
	return
}
