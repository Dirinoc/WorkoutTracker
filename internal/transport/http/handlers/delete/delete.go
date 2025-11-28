package delete

import "time"

type Request struct {
	date time.Time
}

type Response struct {
	WorkoutID int `json:"WorkoutID,omitempty"`
}

func DeleteWorkout(req Request) {
	const op = "handlers.workout.delete.DeleteWorkout"

}
