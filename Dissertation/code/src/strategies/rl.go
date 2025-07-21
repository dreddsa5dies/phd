package strategies

import "github.com/dreddsa5dies/phd/Dissertation/code/src/models"

type RLStrategy struct {
	Rewards map[string]float64
}

func (s *RLStrategy) ChooseTask(robot *models.Robot, tasks []*models.Task, step int) *models.Task {
	var bestTask *models.Task
	maxScore := -1.0

	for _, task := range tasks {
		if task.RequiredEquipment == robot.Equipment && robot.Energy >= task.EnergyCost {
			score := s.Rewards[task.ID] + float64(task.Priority)
			if score > maxScore {
				maxScore = score
				bestTask = task
			}
		}
	}

	if bestTask != nil {
		s.Rewards[bestTask.ID] += 0.1
	}

	return bestTask
}
