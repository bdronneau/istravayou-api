package strava

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bdronneau/istravayou/pkg/models"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	stravaSDK "github.com/strava/go.strava"
)

func (a app) handleLogin(c echo.Context) error {
	toClick := fmt.Sprintf(`<a href="%s">Click Me</a>`, a.authenticator.AuthorizationURL("state1", "activity:read", true))

	return c.HTML(http.StatusOK, "Login! "+toClick)
}

func (a app) handleOAuth(c echo.Context) error {
	if c.Request().FormValue("error") == "access_denied" {
		return c.String(http.StatusUnauthorized, "No access for Login page!")
	}

	t := new(tokenURL)
	if err := c.Bind(t); err != nil {
		return c.String(400, "Nope 400")
	}

	return c.Redirect(302, "/private/info")
}

func (a app) handleInfo(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

func (a app) getAccessToken(code string) (*stravaSDK.AuthorizationResponse, error) {
	client := http.DefaultClient

	data, err := a.authenticator.Authorize(code, client)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (a app) handleAthlete(c echo.Context) error {
	code := c.Request().Header.Get("X-Athlete-Code")

	logrus.Debug(code)
	// TODO: handle better for validation and use common function
	if code == "" || code == "undefined" {
		return c.JSON(400, "No Header X-Athlete-Code")
	}

	data, err := a.getAccessToken(code)

	if err != nil {
		logrus.Errorf("handleAthlete %v", err)
		return c.JSON(401, "Check logs")
	}

	client := stravaSDK.NewClient(data.AccessToken)

	athlete, err := stravaSDK.NewCurrentAthleteService(client).Get().Do()

	if err != nil {
		logrus.Errorf("handleAthlete %v", err)
		return c.JSON(400, "Check logs")
	}

	return c.JSON(200, athlete)

}

// TODO: do not keep this
func oAuthSuccess(auth *stravaSDK.AuthorizationResponse, w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("SUCCESS:\nAt this point you can use this information to create a new user or link the account to one of your existing users\n")
	logrus.Debugf("State: %s\n\n", auth.State)
	logrus.Debugf("Access Token: %s\n\n", auth.AccessToken)

	content, _ := json.MarshalIndent(auth.Athlete, "", " ")

	_, _ = fmt.Fprint(w, string(content))
}

// TODO: do not keep this
func oAuthFailure(err error, w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("Authorization Failure:\n")

	// some standard error checking
	if err == stravaSDK.OAuthAuthorizationDeniedErr {
		logrus.Debug("The user clicked the 'Do not Authorize' button on the previous page.\n")
		logrus.Debug("This is the main error your application should handle.")
	} else if err == stravaSDK.OAuthInvalidCredentialsErr {
		logrus.Debug("You provided an incorrect client_id or client_secret.\nDid you remember to set them at the begininng of this file?")
	} else if err == stravaSDK.OAuthInvalidCodeErr {
		logrus.Debug("The temporary token was not recognized, this shouldn't happen normally")
	} else if err == stravaSDK.OAuthServerErr {
		logrus.Debug("There was some sort of server error, try again to see if the problem continues")
	} else {
		logrus.Debugf("oAuthFailure %v", err)
	}

	_, _ = fmt.Fprint(w, "Check logs")
}

func (a app) handleHeadAthlete(c echo.Context) error {
	code := c.Request().Header.Get("X-Athlete-Code")

	// TODO: handle better for validation and use common function
	if code == "" || code == "undefined" {
		return c.JSON(400, "No Header X-Athlete-Code")
	}

	_, err := a.checkCodeExist(code)

	if err != nil {
		logrus.Errorf("handleHeadAthlete %v", err)
		return c.JSON(401, "Check logs")
	}

	return c.JSON(200, nil)
}

func (a app) handleAuth(c echo.Context) error {
	code := c.Request().Header.Get("X-Athlete-Code")

	// TODO: handle better for validation
	if code == "" || code == "undefined" {
		return c.JSON(400, "No Header X-Athlete-Code")
	}

	auth, err := a.getAccessToken(code)

	if err != nil {
		logrus.Errorf("handleAuth %v", err)
		return c.JSON(400, "Check logs")
	}

	athlete, err := models.GetAthleteByStravaID(auth.Athlete.Id)

	if err == sql.ErrNoRows {
		// TODO: format as models.Athlete
		athlete, err := models.InsertAthlete(auth, code)

		if err != nil {
			logrus.Errorf("handleAuth %v", err)
			return c.JSON(500, "Check logs")
		}

		return c.JSON(201, athlete)
	} else if err != nil {
		logrus.Errorf("handleAuth %v", err)
		return c.JSON(401, "Check logs")
	}

	athlete.Code = code

	// TODO: Handle error
	_, _ = models.UpdateAthleteCode(athlete)

	return c.JSON(200, nil)
}

func (a app) checkCodeExist(code string) (*models.Athlete, error) {
	athlete, err := models.GetAthleteByCode(code)

	if err != nil {
		return nil, err
	}

	return athlete, nil
}
