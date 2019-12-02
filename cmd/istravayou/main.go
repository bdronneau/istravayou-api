package main

import (
	"flag"
	"log"
	"os"

	"github.com/bdronneau/istravayou/pkg/strava"

	"github.com/peterbourgon/ff"
	"github.com/sirupsen/logrus"
)

func main() {
	fs := flag.NewFlagSet("istravayou", flag.ExitOnError)

	var (
		loglevel    = fs.String("log-level", "DEBUG", "Log level")
		environment = fs.String("env", "prod", "run level for app")
	)

	stravaConfig := strava.Flags(fs)

	err := ff.Parse(fs, os.Args[1:],
		ff.WithConfigFileFlag("config"),
		ff.WithConfigFileParser(ff.PlainParser),
		ff.WithEnvVarPrefix("ISTRAVAYOU"),
	)

	if err != nil {
		log.Fatalf("error while parsing flags: %v", err)
	}

	if err := configLogger(*loglevel); err != nil {
		log.Fatalf("unable to configure logger: %v", err)
	}

	app, _ := strava.New(stravaConfig)

	app.NewHTTP(*environment)
}

// TODO: Go to helpers
func configLogger(logLevel string) error {
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.LUTC)

	logrusLevel, err := logrus.ParseLevel(logLevel)
	if err != nil {
		return err
	}

	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrusLevel)

	return nil
}
