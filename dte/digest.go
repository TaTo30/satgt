package dte

import (
	"log"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

func AddWorkload(app core.App, files []DTEFile) {
	for _, file := range files {
		collection, err := app.FindCollectionByNameOrId("digest_workload")
		if err != nil {
			log.Println(err.Error())
			return
		}

		workload, err := app.FindFirstRecordByData("digest_workload", "uuid", file.Filename)
		if err == nil {
			// The file is already stored in digest, skip it
			log.Printf("File %s already store in digest\n", file.Filename)
			continue
		}

		workload = core.NewRecord(collection)
		content := string(file.Bytes)

		workload.Set("uuid", file.Filename)
		workload.Set("content", content)
		workload.Set("state", "P")

		err = app.Save(workload)
		if err != nil {
			log.Println(err.Error())
			return
		}
	}
}

func ProcessWorkloads(app core.App, state string) {
	if state == "" {
		state = "P"
	}

	records, err := app.FindRecordsByFilter("digest_workload", "state = {:state}", "", 0, 0, dbx.Params{"state": state})
	if err != nil {
		log.Println(err.Error())
		return
	}

	for _, record := range records {
		content := record.Get("content")
		contentbytes := []byte(content.(string))
		dtexml, err := ParseXMLFile(contentbytes)

		if err != nil {
			record.Set("message", err.Error())
			record.Set("state", "E")

			if errs := app.Save(record); errs != nil {
				log.Println(err.Error())
				return
			}

			continue
		}

		if err := AddDTE(app, dtexml); err != nil {
			record.Set("message", err.Error())
			record.Set("state", "E")

			if errs := app.Save(record); errs != nil {
				log.Println(err.Error())
				return
			}

			continue
		}

		record.Set("state", "C")
		err = app.Save(record)

		if err != nil {
			log.Println(err.Error())
			return
		}
	}
}
