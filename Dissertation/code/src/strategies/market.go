package strategies

import "github.com/dreddsa5dies/phd/Dissertation/code/src/models"

type MarketStrategy struct{}

func (s *MarketStrategy) ChooseTask(robot *models.Robot, tasks []*models.Task, step int) *models.Task {
	var bestTask *models.Task
	maxBid := -1.0

	for _, task := range tasks {
		if task.RequiredEquipment == robot.Equipment && robot.Energy >= task.EnergyCost {
			bid := float64(task.Priority) / task.EnergyCost
			if bid > maxBid {
				maxBid = bid
				bestTask = task
			}
		}
	}
	return bestTask
}
