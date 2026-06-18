package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"satgt/dte"
	_ "satgt/migrations"
	"satgt/session"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"github.com/pocketbase/pocketbase/tools/osutils"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	// Set environment variables
	processWorkloadCron := os.Getenv("CRON_PROCESS_WORKLOADS")
	if processWorkloadCron == "" {
		processWorkloadCron = "*/5 * * * *"
	}

	app := pocketbase.New()

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.GET("/satgt/fel", func(re *core.RequestEvent) error {
			sessionFetcher := &session.SatSSOFEL{}
			session, err := sessionFetcher.FetchSession(false)

			if err != nil {
				re.JSON(400, err.Error())
			}

			return re.JSON(200, session.AccessToken)
		})
		return se.Next()
	})

	// Enable this function to download data from official SAT portal
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.GET("/satgt/dte", func(re *core.RequestEvent) error {
			execute := func(year int, month int, opType string, session *session.SatSession) int {
				log.Printf("Fetching data for %d/%d\n", month, year)
				params := dte.FetchDTEParameters(month, year, opType)
				result, err := dte.FetchDTEData(params, *session)
				if err != nil {
					log.Println(err.Error())
					return 0
				}

				file, err := dte.FetchDTEXMLFiles(result.Detail.Data, params, *session)
				if err != nil {
					log.Println(err.Error())
					return 0
				}

				dtefiles, err := dte.ExtractXMLFilesFromZip(file)
				if err != nil {
					log.Println(err.Error())
					return 0
				}

				dte.AddWorkload(app, dtefiles)
				return len(dtefiles)
			}

			queryparams := re.Request.URL.Query()
			opType := queryparams.Get("type")
			if opType != "R" && opType != "E" {
				log.Printf("Operation type invalid (%s), falling back to 'R'", opType)
				opType = "R"
			}

			month, noMonth := strconv.ParseInt(queryparams.Get("month"), 10, 32)
			year, err := strconv.ParseInt(queryparams.Get("year"), 10, 32)
			if err != nil {
				year = int64(time.Now().Year())
			}

			log.Printf("Retreiving data for year %d and month-set %t, mode: %s\n", year, noMonth != nil, opType)
			sessionFetcher := &session.SatSSODTE{}
			session, err := sessionFetcher.FetchSession(false)
			if err != nil {
				log.Println(err.Error())
				return re.String(500, err.Error())
			}

			filesCount := 0
			if noMonth != nil {
				// If user did not specify a month let's fetch all data from january to current month
				// if the year set is the current year, if the year is past retreive data of all months
				currYear := time.Now().Year()
				currMonth := time.Now().Month()
				if year < int64(currYear) {
					currMonth = 12
				}

				for m := 1; m <= int(currMonth); m++ {
					filesCount += execute(int(year), int(m), opType, session)
				}
			} else {
				// If user specified a month, let's fetch that month for the selected year
				filesCount += execute(int(year), int(month), opType, session)
			}

			return re.JSON(200, fmt.Sprintf("%d DTE documents fetched", filesCount))
		})

		return se.Next()
	})

	app.Cron().MustAdd("process_workloads", processWorkloadCron, func() {
		dte.ProcessWorkloads(app, "P")
	})

	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		e.UIExtensions = append(e.UIExtensions, core.UIExtension{
			Name: "satgt",
			FS:   os.DirFS("./ui"),
		})

		return e.Next()
	})

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: osutils.IsProbablyGoRun(),
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
