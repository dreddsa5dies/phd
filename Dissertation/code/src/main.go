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
		&strategies.RLStrategy{},
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
			Energy:    rc.Energy,
		}
	}

	// инициализация задач
	tasks := make([]models.Task, len(config.Tasks))
	for i, tc := range config.Tasks {
		tasks[i] = models.Task{
			ID:                tc.ID,
			RequiredEquipment: tc.RequiredEquipment,
			EnergyCost:        tc.EnergyCost,
			CreatedAt:         time.Now().Unix(),
		}
	}

	// добавление задач в массив
	// наблюдаемых машинами (случайный процесс)
	tasks = tasksFromMachines(
		config.Simulation.PoissonLambda,
		tasks,
		config.Simulation.Steps,
	)

	// прогон всех стратегий на одинаковых исходных данных
	// каждая стратегия прогоняется заданное количество шагов
	// формируется отчет
	results := make(map[string]strategies.Metrics)
	for _, s := range allStrategies {
		fmt.Printf("Старт эксперимента со стратегией: %s\n", s)

		// Копируем список задач и машин
		// независимые прогоны стратегий
		taskList := make([]models.Task, len(tasks))
		copy(taskList, tasks)

		machineList := make([]models.Machine, len(machines))
		copy(machineList, machines)

		engine.RunSimulation(
			machineList,
			taskList,
			s,
			config.Simulation.Steps,
			results,
		)
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
	rand.New(rand.NewSource(time.Now().UnixNano()))

	for range steps {
		if rand.Float64() < lambda {
			// всего 13 типов в этой симуляции
			eq := 1 + rand.Intn(12) //nolint:mnd
			// определим энергоемкость задачи не больше 50
			en := 1 + rand.Intn(49) //nolint:mnd

			newTask := models.Task{
				ID:                fmt.Sprintf("T%d", len(tasks)+1),
				RequiredEquipment: models.TypeEquipment(eq),
				EnergyCost:        float64(en),
				CreatedAt:         time.Now().Unix(),
			}
			tasks = append(tasks, newTask)
		}
	}

	return tasks
}
