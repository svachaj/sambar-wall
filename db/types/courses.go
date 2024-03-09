package types

import "time"

type Course struct {
	ID                int       `json:"id" db:"id"`
	Name              string    `json:"name" db:"name"`
	Description       string    `json:"description" db:"description"`
	Days              string    `json:"days" db:"days"`
	AgeGroup          string    `json:"ageGroup" db:"ageGroup"`
	Capacity          int       `json:"capacity" db:"capacity"`
	ApplicationsCount int       `json:"applicationsCount" db:"applicationsCount"`
	TimeFrom          time.Time `json:"valid_from" db:"timeFrom"`
	TimeTo            time.Time `json:"valid_to" db:"timeTo"`
	PartipicatnsCount int       `json:"partipicatnsCount" db:"partipicatnsCount"`
	Price             float64   `json:"price" db:"price"`
	DurationMin       int       `json:"durationMin" db:"durationMin"`
}

type CourseType struct {
	ID          int      `json:"id" db:"id"`
	Name        string   `json:"name" db:"name"`
	Description string   `json:"description" db:"description"`
	Courses     []Course `json:"courses" db:"courses"`
}
