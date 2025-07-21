package strategies

import "github.com/dreddsa5dies/phd/Dissertation/code/src/models"

type Strategy interface {
	ChooseTask(robot *models.Robot, tasks []*models.Task, step int) *models.Task
}
