package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/dreddsa5dies/phd/Dissertation/code/src/engine"
	"github.com/dreddsa5dies/phd/Dissertation/code/src/models"
	"github.com/dreddsa5dies/phd/Dissertation/code/src/strategies"
	"github.com/dreddsa5dies/phd/Dissertation/code/src/utils"
)

func main() {
	// чтение конфигурационного файла
	config, err := utils.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	// инициализация стратегий поведения
	allStrategies := []strategies.Strategy{
		// "Стратегия прямого перебора"
		&strategies.DirectSearch{},
		// "Рыночная стратегия"
		&strategies.MarketStrategy{},
		// "RL"
		&strategies.RLStrategy{
			Q:       make(map[string]map[int]float64),
			LR:      0.5,
			Gamma:   0.9,
			Epsilon: 0.2,
			MinEps:  0.01,
			Decay:   0.995,
		},
		// "Энергосберегающая стратегия"
		&strategies.EnergyEfficientStrategy{},
		// "Стратегия минимизации временных задержек"
		&strategies.DelayMinimizationStrategy{
			// коэффициенты для
			Alpha: 0.5, //nolint:mnd
			Beta:  1.0,
		},
	}

	// инициализация группы машин
	machines := make([]models.Machine, len(config.Machines))
	for i, rc := range config.Machines {
		machines[i] = models.Machine{
			ID:        rc.ID,
			Equipment: rc.Equipment,
		}
	}

	// инициализация задач
	tasks := make([]models.Task, len(config.Tasks))
	tNow := time.Now().Unix()
	for i, tc := range config.Tasks {
		tasks[i] = models.Task{
			ID:                tc.ID,
			RequiredEquipment: tc.RequiredEquipment,
			EnergyCost:        tc.EnergyCost,
			CreatedAt:         tNow,
		}

		// обеспечение временной последовательности
		tNow++
	}

	// добавление задач в массив
	// наблюдаемых машинами (случайный процесс)
	tasks = tasksFromMachines(
		config.Simulation.PoissonLambda,
		tasks,
		config.Simulation.Steps,
	)

	// перемешивание задач в массиве случайным образом
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(tasks), func(i, j int) {
		tasks[i], tasks[j] = tasks[j], tasks[i]
	})

	// прогон всех стратегий на одинаковых исходных данных
	// каждая стратегия прогоняется заданное количество шагов
	// формируется отчет
	results := make(map[string]map[string]strategies.Metrics)

	// Вычисляем суммарную энергию всех задач
	var totalTaskEnergy float64
	for _, task := range tasks {
		totalTaskEnergy += task.EnergyCost
	}

		// Определяем три сценария по энергии машин
		scenarios := map[string]float64{
			// СУМ[энергия всех НТТС] = СУМ[энергия всех задач] / 2 -> min
			"min": totalTaskEnergy / 2.0,
			// СУМ[энергия всех НТТС] = СУМ[энергия всех задач] -> equal
			"equal": totalTaskEnergy,
			// СУМ[энергия всех НТТС] = СУМ[энергия всех задач] * 2 -> max
			"max": totalTaskEnergy * 2.0,
		}

	// Восстанавливаем исходные машины (до модификации энергии)
	originalMachinesConfig := config.Machines

	for scenarioName, totalMachineEnergy := range scenarios {
		fmt.Printf("\n+++ Сценарий: %s (суммарная энергия машин = %.2f)\n", scenarioName, totalMachineEnergy)

		// Создаём новые машины с той же структурой, но общей энергией, равной сценарию
		machines := make([]models.Machine, len(originalMachinesConfig))

		// Распределяем totalMachineEnergy равномерно между машинами
		energyPerMachine := totalMachineEnergy / float64(len(machines))
		for i, rc := range originalMachinesConfig {
			machines[i] = models.Machine{
				ID:        rc.ID,
				Equipment: rc.Equipment,
				Energy:    energyPerMachine, // равномерное распределение
			}
		}

		// Инициализация под-отчёта для этого сценария
		results[scenarioName] = make(map[string]strategies.Metrics)

		// Запуск всех стратегий в этом сценарии
		for _, s := range allStrategies {
			strategyName := fmt.Sprintf("%s", s)
			fmt.Printf("Старт эксперимента со стратегией: %s\n", strategyName)

			// Копируем задачи, чтобы не мутировать оригинал
			taskCopies := make([]models.Task, len(tasks))
			copy(taskCopies, tasks)

			// Копируем машины, чтобы не мутировать оригинал
			machineCopies := make([]models.Machine, len(machines))
			copy(machineCopies, machines)

			engine.RunSimulation(
				machineCopies,
				taskCopies,
				s,
				config.Simulation.Steps,
				results[scenarioName], // передаём под-отчёт
			)
		}
	}

	// сохранение результатов
	err = engine.GenerateReport(results)
	if err != nil {
		fmt.Printf("Ошибка сохранения отчета: %s\n", err)
		return
	}

	fmt.Println("Эксперименты закончены")
}

// случайный процесс добавления новых задач, наблюдаемых машинами
// один для ВСЕХ стратегий
func tasksFromMachines(lambda float64, tasks []models.Task, steps int) []models.Task {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	tNow := time.Now().Unix()
	for range steps {
		if r.Float64() < lambda {
			// всего 15 типов в этой симуляции
			eq := 1 + r.Intn(14) //nolint:mnd
			// определим энергоемкость задачи: максимум не больше 3 * 1000
			en := 1 + r.Intn(2999) //nolint:mnd

			newTask := models.Task{
				ID:                fmt.Sprintf("T%d", len(tasks)+1),
				RequiredEquipment: models.TypeEquipment(eq),
				EnergyCost:        float64(en),
				CreatedAt:         tNow,
			}
			tasks = append(tasks, newTask)

			// обеспечение временной последовательности
			tNow++
		}
	}

	return tasks
}
