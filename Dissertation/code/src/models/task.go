package models

type Task struct {
	ID                string
	RequiredEquipment string
	EnergyCost        float64
	Priority          int
	Deadline          int
	CreatedAt         int
}
