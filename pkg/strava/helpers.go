package strava

import (
	"github.com/bdronneau/istravayou/pkg/models"
	"github.com/sirupsen/logrus"
)

func (a app) checkCodeExist(code string) (*models.Athlete, error) {
	athlete, err := models.GetAthleteByCode(code)

	if err != nil {
		logrus.Errorf("When retrieve athlete by code %v", err)
		return nil, err
	}

	return athlete, nil
}

func (a app) getAccessToken(code string) (*models.Athlete, error) {
	data, err := models.GetAthleteByCode(code)

	if err != nil {
		return nil, err
	}

	return data, nil
}
