package engine

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/dreddsa5dies/phd/Dissertation/code/src/models"
	"github.com/dreddsa5dies/phd/Dissertation/code/src/strategies"
	"github.com/dreddsa5dies/phd/Dissertation/code/src/utils"
)

type SimulationResult struct {
	StrategyName    string
	TotalTasksDone  int
	TotalEnergyUsed float64
	AvgTaskDelay    float64
}

func RunSimulation(config *utils.Config, strategy strategies.Strategy) SimulationResult {
	rand.Seed(time.Now().UnixNano())

	robots := make([]*models.Robot, len(config.Robots))
	for i, rc := range config.Robots {
		robots[i] = &models.Robot{
			ID:        rc.ID,
			Equipment: rc.Equipment,
			Energy:    rc.Energy,
		}
	}

	tasks := make([]*models.Task, len(config.Tasks))
	for i, tc := range config.Tasks {
		tasks[i] = &models.Task{
			ID:                tc.ID,
			RequiredEquipment: tc.RequiredEquipment,
			EnergyCost:        tc.EnergyCost,
			Priority:          tc.Priority,
			Deadline:          tc.Deadline,
			CreatedAt:         0,
		}
	}

	totalTasksDone := 0
	totalEnergyUsed := 0.0

	for step := 0; step < config.Simulation.MaxSteps; step++ {
		if rand.Float64() < config.Simulation.PoissonLambda {
			newTask := &models.Task{
				ID:                fmt.Sprintf("T%d", len(tasks)+1),
				RequiredEquipment: "loader",
				EnergyCost:        10,
				Priority:          2,
				Deadline:          5,
				CreatedAt:         step,
			}
			tasks = append(tasks, newTask)
		}

		for _, robot := range robots {
			if robot.AssignedTask == nil {
				chosenTask := strategy.ChooseTask(robot, tasks, step)
				if chosenTask != nil {
					robot.AssignedTask = chosenTask
					robot.Energy -= chosenTask.EnergyCost
					totalEnergyUsed += chosenTask.EnergyCost
					tasks = removeTask(tasks, chosenTask)
					totalTasksDone++
				}
			}
		}

		for _, robot := range robots {
			if robot.AssignedTask != nil {
				fmt.Printf("Step %d: Robot %s is doing task %s\n", step, robot.ID, robot.AssignedTask.ID)
				robot.AssignedTask = nil
			}
		}
	}

	return SimulationResult{
		StrategyName:    "Unknown",
		TotalTasksDone:  totalTasksDone,
		TotalEnergyUsed: totalEnergyUsed,
		AvgTaskDelay:    0,
	}
}

func removeTask(tasks []*models.Task, task *models.Task) []*models.Task {
	for i, t := range tasks {
		if t.ID == task.ID {
			return append(tasks[:i], tasks[i+1:]...)
		}
	}
	return tasks
}
