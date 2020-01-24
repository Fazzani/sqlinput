package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	. "github.com/fazzani/sqlinput/models"
	"github.com/jackc/pgx/v4"
	"github.com/spf13/cast"
)

func doScheme() {
	print(SCHEME)
}

// Empty validation routine. This routine is optional.
func validateArguments() {

}

func getCols(rows *pgx.Rows) []string {
	fieldDescriptions := (*rows).FieldDescriptions()
	var columns []string
	for _, col := range fieldDescriptions {
		columns = append(columns, string(col.Name))
	}

	return columns
}

func getTemplatedQuery(lastID string, query *string) (err error) {
	t := template.Must(template.New("").Parse(*query))
	data := map[string]string{
		"checkpoint": lastID,
	}
	var buf bytes.Buffer

	if err = t.Execute(&buf, data); err != nil {
		return
	}

	*query = buf.String()
	return
}

func runScript(sqlInputConfig *SQLInputConfig, conf *ConfigInput) (err error) {

	ctx := context.Background()

	Logf(LogDebug, "connString: %s", sqlInputConfig.ConnString)
	Logf(LogDebug, "query: %s", sqlInputConfig.Query)
	// Logf(LogInfo, "%s", conf)

	checkpointPath := conf.GetCheckpointPath()
	Logf(LogInfo, "checkpointPath: %s sqlInputConfig.Checkpoint: %v", checkpointPath, sqlInputConfig.Checkpoint)

	if _, err = os.Stat(checkpointPath); sqlInputConfig.Checkpoint && err == nil {
		Logf(LogDebug, "checkpoint enabled")
		lastID := sqlInputConfig.CheckpointIDStart
		if lastID, err = conf.GetLastCheckpointValue(sqlInputConfig.CheckpointIDStart); err != nil {
			Logf(LogError, "%v", err)
		}

		Logf(LogInfo, "lastID: %s", lastID)

		if err = getTemplatedQuery(lastID, &sqlInputConfig.Query); err != nil {
			Logf(LogError, "%v", err)
			return
		}

		Logf(LogDebug, "Query after template rendering: %s\n", sqlInputConfig.Query)
	} else {
		if sqlInputConfig.Checkpoint {
			Logf(LogWarn, "Checkpoint file not yet exist")
			if err = getTemplatedQuery(sqlInputConfig.CheckpointIDStart, &sqlInputConfig.Query); err != nil {
				Logf(LogError, "%v", err)
				return
			}
		}
	}

	conn, err := pgx.Connect(ctx, sqlInputConfig.ConnString)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	Logf(LogInfo, "Database successful connection")

	rows, err := conn.Query(ctx, sqlInputConfig.Query)
	if err != nil {
		return err
	}
	defer rows.Close()

	columns := ([]string)(nil)

	for rows.Next() {
		now := time.Now()
		plainRow := fmt.Sprintf("_time=%s, timestamp=%d, enviroment=%s", now.Format("Mon, 2 Jan 2006 15:04:05 MST"), time.Now().Unix(), sqlInputConfig.Environment)
		rawValues, _ := rows.Values()
		if columns == nil {
			columns = getCols(&rows)
			Logf(LogDebug, strings.Join(columns, ", "))
		}

		for i, col := range columns {
			// val := cast.ToString(rawValues[i])
			plainRow += fmt.Sprintf(", %s=%s", col, cast.ToString(rawValues[i]))
			// if strings.ContainsAny(val, " ") {
			// 	plainRow += fmt.Sprintf("%s='%s' ", col, cast.ToString(rawValues[i]))
			// } else {
			// }
		}
		_, err = fmt.Printf("%s\r\n", plainRow)
		time.Sleep(200 * time.Millisecond)
	}

	Logf(LogInfo, "Save checkpoint")
	//save checkpoint
	if sqlInputConfig.Checkpoint {
		rowsCP, errCP := conn.Query(ctx, sqlInputConfig.CheckpointIDQuery)
		defer rowsCP.Close()
		if errCP != nil {
			Logf(LogError, "%v", errCP)
		}
		for rowsCP.Next() {
			checkpointID, _ := rowsCP.Values()
			Logf(LogDebug, "the newest checkpoint id is: %v for the CheckpointID Query: %s", cast.ToString(checkpointID[0]), sqlInputConfig.CheckpointIDQuery)
			err = conf.SaveCheckpoint(cast.ToString(checkpointID[0]))
			if err != nil {
				Logf(LogError, "%v", err)
			}
		}
	}

	if err != nil {
		return err
	}
	return nil
}

func main() {

	env := os.Getenv("ENVIRONMENT")
	isLocal := strings.EqualFold(env, "local")

	scheme := flag.Bool("scheme", false, "scheme")
	validateArgs := flag.Bool("validate-arguments", false, "validate arguments")

	// Logf(LogInfo, "Enviroment local=%v", isLocal)

	flag.Parse()

	if *scheme {
		doScheme()
	} else if *validateArgs {
		validateArguments()
	} else {
		conf := &ConfigInput{}
		if err := conf.Get(); err != nil {
			Logf(LogError, "Reading configuration failed: %v\n", err.Error())
			os.Exit(1)
		}
		// Logf(LogDebug, "config input: %v", conf)

		sqlInputConfig := &SQLInputConfig{"", "", env, false, "", ""}

		if isLocal {
			// Get configuration for local dev
			sqlInputConfig.Query = "SELECT * FROM volumes_staging.expected_feeds WHERE id > {{.checkpoint}}"
			sqlInputConfig.ConnString = "host='localhost' dbname='datatrust' user='dt_user' password='dtpass'"
			sqlInputConfig.CheckpointIDQuery = "SELECT MAX(id) FROM volumes_staging.expected_feeds"
			sqlInputConfig.CheckpointIDStart = "0"
			sqlInputConfig.Checkpoint = true
		} else {
			// Get configuration from Splunk
			Logf(LogInfo, "Getting configuration from Splunk")
			for _, e := range conf.Configuration.Stanza.Param {
				switch e.Name {
				case "query":
					sqlInputConfig.Query = e.Text
				case "connectionstring":
					sqlInputConfig.ConnString = e.Text
				case "environment":
					sqlInputConfig.Environment = e.Text
				case "checkpoint":
					sqlInputConfig.Checkpoint = cast.ToBool(e.Text)
				case "checkpoint_id_query":
					sqlInputConfig.CheckpointIDQuery = e.Text
				case "checkpoint_id_start":
					sqlInputConfig.CheckpointIDStart = e.Text
				}
			}
		}

		Logf(LogInfo, "SQLInput configuration %v", sqlInputConfig)

		err := runScript(sqlInputConfig, conf)
		if err != nil {
			Logf(LogError, "Query failed: %v\n", err.Error())
			os.Exit(1)
		}
	}

	os.Exit(0)
}
