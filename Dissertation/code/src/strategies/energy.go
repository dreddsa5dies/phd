package strategies

import (
	"math"

	"github.com/dreddsa5dies/phd/Dissertation/code/src/models"
)

type EnergyEfficientStrategy struct{}

func (s *EnergyEfficientStrategy) ChooseTask(robot *models.Robot, tasks []*models.Task, step int) *models.Task {
	var bestTask *models.Task
	minCost := math.MaxFloat64
	for _, task := range tasks {
		if task.RequiredEquipment == robot.Equipment && robot.Energy >= task.EnergyCost {
			if task.EnergyCost < minCost {
				minCost = task.EnergyCost
				bestTask = task
			}
		}
	}
	return bestTask
}
