package mysql

import (
	"database/sql"
	"fmt"
	"log"

	npq "github.com/Knetic/go-namedParameterQuery"
	_ "github.com/go-sql-driver/mysql"

	"DeadRabbit/state"
)

type Configuration struct {
	Host     string
	Port     string
	User     string
	Password string
	Schema   string
}

type DatabaseQueryResults struct {
	headers []string
	rows    []map[string]string
}

func (d DatabaseQueryResults) GetHeaders() []string {
	return d.headers
}

func (d DatabaseQueryResults) GetResults() []map[string]string {
	return d.rows
}

type Database struct {
	db *sql.DB
}

func New(c Configuration) (Database, error) {
	connectionStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", c.User, c.Password, c.Host, c.Port, c.Schema)

	db, err := sql.Open("mysql", connectionStr)
	if err != nil {
		return Database{}, err
	}
	err = db.Ping()
	if err != nil {
		return Database{}, err
	}

	return Database{
		db: db,
	}, nil
}

func (d *Database) Query(query string, params map[string]string) state.QueryResults {
	if d.db == nil {
		log.Fatal("DB Connection isn't created")
	}

	aQuery := npq.NewNamedParameterQuery(query)
	for key, value := range params {
		aQuery.SetValue(key, value)
	}

	log.Printf("Executing query \"%s\"\nParameters: %+v\n", aQuery.GetParsedQuery(), aQuery.GetParsedParameters())
	rows, err := d.db.Query(aQuery.GetParsedQuery(), aQuery.GetParsedParameters()...)
	if err != nil {
		log.Fatalf("Can't execute statement, err: %v", err)
	}
	defer rows.Close()

	headers, _ := rows.Columns()

	parsedRows := make([]map[string]string, 0)

	for rows.Next() {
		destinations := make([]any, 0)
		for i := 0; i < len(headers); i++ {
			aStringDest := ""
			destinations = append(destinations, &aStringDest)
		}

		err = rows.Scan(destinations...)
		if err != nil {
			log.Fatalf("Can't parse results: %v", err)
		}
		parsedRow := map[string]string{}
		for i, header := range headers {
			parsedRow[header] = fmt.Sprintf("%v", *(destinations[i].(*string)))
		}
		parsedRows = append(parsedRows, parsedRow)
	}

	return DatabaseQueryResults{
		headers: headers,
		rows:    parsedRows,
	}
}

func (d *Database) Close() {
	_ = d.db.Close()
}
