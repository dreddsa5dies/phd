package strategies

import "github.com/dreddsa5dies/phd/Dissertation/code/src/models"

type DelayMinimizationStrategy struct{}

func (s *DelayMinimizationStrategy) ChooseTask(robot *models.Robot, tasks []*models.Task, step int) *models.Task {
	var bestTask *models.Task
	highestPriority := -1
	for _, task := range tasks {
		if task.RequiredEquipment == robot.Equipment && robot.Energy >= task.EnergyCost {
			if task.Priority > highestPriority {
				highestPriority = task.Priority
				bestTask = task
			}
		}
	}
	return bestTask
}
