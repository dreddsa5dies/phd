package models

type Robot struct {
	ID           string
	Equipment    string
	Energy       float64
	AssignedTask *Task
}
