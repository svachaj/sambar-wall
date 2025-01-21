package types

import "time"

type Course struct {
	ID                int       `json:"id" db:"id"`
	Name              string    `json:"name" db:"name"`
	Code              string    `json:"code" db:"code"`
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

type ApplicationForm struct {
	ID             int        `json:"id" db:"id"`
	FirstName      string     `json:"firstName" db:"firstName"`
	LastName       string     `json:"lastName" db:"lastName"`
	PersonalID     *string    `json:"personalId" db:"personalId"`
	BirthYear      *int       `json:"birthYear" db:"birthYear"`
	HealthState    *string    `json:"healthState" db:"healthState"`
	Paid           bool       `json:"gdpr" db:"paid"`
	CourseID       int        `json:"courseId" db:"courseId"`
	CourseName     string     `json:"courseName" db:"courseName"`
	CourseCode     string     `json:"courseCode" db:"courseCode"`
	CourseDays     string     `json:"courseDays" db:"courseDays"`
	CourseTimeFrom time.Time  `json:"courseTimeFrom" db:"courseTimeFrom"`
	CourseTimeTo   time.Time  `json:"courseTimeTo" db:"courseTimeTo"`
	CourseAgeGroup string     `json:"courseAgeGroup" db:"courseAgeGroup"`
	CoursePrice    float64    `json:"coursePrice" db:"coursePrice"`
	Email          *string    `json:"email" db:"email"`
	Phone          *string    `json:"phone" db:"phone"`
	ParentName     *string    `json:"parentName" db:"parentName"`
	CreatedDate    *time.Time `json:"createdDate" db:"createdDate"`
	ParticipantID  *int       `json:"participantId" db:"participantId"`
	IsActive       *bool      `json:"isActive" db:"isActive"`
	WillContinue   *bool      `json:"willContinue" db:"willContinue"`
	CreatedByID    int        `json:"createdById" db:"createdById"`
}
