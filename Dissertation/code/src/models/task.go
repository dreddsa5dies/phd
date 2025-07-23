package models

// Task описывает задачу
type Task struct {
	// уникальный идентификатор
	ID string `yaml:"id"`
	// требуемый тип оборудования
	RequiredEquipment TypeEquipment `yaml:"required_equipment"`
	// требуемые затраты по энергии
	EnergyCost float64 `yaml:"energy_cost"`
	// время создания в UnixTime
	CreatedAt int64
}

// удаление задачи из массива
func RemoveTask(tasks []Task, task Task) []Task {
	for i, t := range tasks {
		if t.ID == task.ID {
			return append(tasks[:i], tasks[i+1:]...)
		}
	}
	return tasks
}

func TotalEnergyTasks(in []Task) float64 {
	total := 0
	for i := range in {
		total += int(in[i].EnergyCost)
	}
	return float64(total)
}
