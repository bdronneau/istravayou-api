package models

import (
	"database/sql"
	"time"

	strava "github.com/bdronneau/go.strava"
	"github.com/sirupsen/logrus"
)

// Athlete is the representation of an athlete
type Athlete struct {
	ID           uint32 `json:"id" form:"id" query:"id"`
	StravaID     int64  `db:"strava_id" json:"strava_id" form:"strava_id" query:"strava_id"`
	Name         string `json:"name" form:"name" query:"name"`
	Code         string `json:"code" form:"code" query:"code"`
	AccessToken  string `db:"access_token"`
	RefreshToken string `db:"refresh_token"`
	Raw          strava.AthleteDetailed
	LastUpdated  string `json:"lastupdated" form:"lastupdated" query:"lastupdated"`
}

// GetAthleteByCode retrieve athlete detail
func GetAthleteByCode(code string) (*Athlete, error) {
	row := db.QueryRow("SELECT id, code, access_token, refresh_token FROM athletes WHERE code=$1", code)

	athlete := &Athlete{}

	err := row.Scan(&athlete.ID, &athlete.Code, &athlete.AccessToken, &athlete.RefreshToken)

	if err == sql.ErrNoRows {
		logrus.WithFields(logrus.Fields{
			"code": code,
		}).Debug("No row for this code")
		// Replace by std error not a sql.ErrNoRows
		return nil, err
	} else if err != nil {
		return nil, err
	}

	return athlete, nil
}

// GetAthleteByStravaID retrieve athlete from strava ID
func GetAthleteByStravaID(id int64) (*Athlete, error) {
	row := db.QueryRow("SELECT id, strava_id FROM athletes WHERE strava_id=$1", id)

	athlete := &Athlete{}

	err := row.Scan(&athlete.ID, &athlete.StravaID)

	if err == sql.ErrNoRows {
		logrus.WithField("strava_id", id).Info("No row")
		// Replace by std error not a sql.ErrNoRows
		return nil, err
	} else if err != nil {
		return nil, err
	}

	return athlete, nil
}

// InsertAthlete insert new athlete
func InsertAthlete(athlete *strava.AuthorizationResponse, code string) (*Athlete, error) {
	sqlStatement := `
		INSERT INTO athletes (strava_id, name, code, access_token, refresh_token, lastupdated)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, strava_id`

	athleteDB := &Athlete{}
	timeUpdated, _ := timeIn(time.Now(), "")
	err := db.QueryRow(
		sqlStatement,
		athlete.Athlete.Id,
		athlete.Athlete.AthleteSummary.FirstName,
		code,
		athlete.AccessToken,
		athlete.RefreshToken,
		timeUpdated).Scan(
		&athleteDB.ID,
		&athleteDB.StravaID)

	if err != nil {
		return nil, err
	}

	return athleteDB, nil
}

// UpdateAthleteCode update athlete
func UpdateAthleteCode(athlete *Athlete) (*Athlete, error) {
	contextLogger := logrus.WithFields(logrus.Fields{
		"id":        athlete.ID,
		"strava_id": athlete.StravaID,
		"code":      athlete.Code,
	})
	contextLogger.Debugf("Update athlete")

	sqlStatement := `
		UPDATE athletes
		SET code = $3, lastupdated = $2, access_token = $4, refresh_token = $5
		WHERE id = $1
		RETURNING id, code`

	athleteDB := &Athlete{}
	timeUpdated, _ := timeIn(time.Now(), "")
	err := db.QueryRow(
		sqlStatement,
		athlete.ID,
		timeUpdated,
		athlete.Code,
		athlete.AccessToken,
		athlete.RefreshToken).Scan(
		&athleteDB.ID,
		&athleteDB.Code)

	if err != nil {
		contextLogger.Errorf("On update athlete %d got %v", athlete.ID, err)
		return nil, err
	}

	return athleteDB, nil
}

// TODO: move to helpers
func timeIn(t time.Time, name string) (time.Time, error) {
	loc, err := time.LoadLocation(name)
	if err == nil {
		t = t.In(loc)
	}
	return t, err
}
