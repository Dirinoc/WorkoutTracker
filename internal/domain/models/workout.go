package models

import "time"

type Workout struct {
	ID         int
	UserID     int
	Date       time.Time
	Excercises []Excercise
}
