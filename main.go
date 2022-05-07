package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"gopkg.in/yaml.v2"

	"DeadRabbit/commons"
	"DeadRabbit/layout"
	"DeadRabbit/mysql"
	"DeadRabbit/rabbitmq"
	"DeadRabbit/state"
	"DeadRabbit/store"
)

const configPath = "configuration.yaml"

var (
	aConfiguration configuration
	aStore         *store.Store[state.State]
	aLayout        *layout.Layout
)

type configuration struct {
	Rabbitmq  rabbitmq.Configuration
	Debug     bool
	Databases []struct {
		Host     string
		Port     string
		User     string
		Password string
		Schema   string
		Name     string
		Queries  []struct {
			Format string
			Name   string
			Params []struct {
				Name   string
				Format string
			}
		}
	}
}

type view struct {
	x, y          int
	width, height int
}

func main() {
	initLogger()
	err := loadConfiguration()
	if err != nil {
		log.Fatal("Can't load configuration")
	}

	sqlQueryOptions := make([]state.SelectableOption, 0)
	for _, db := range aConfiguration.Databases {
		database, err := mysql.New(mysql.Configuration{
			Host:     db.Host,
			Port:     db.Port,
			User:     db.User,
			Password: db.Password,
			Schema:   db.Schema,
		})
		if err != nil {
			log.Fatalf("Can't connect to db %s\nConnection settings: %+v\nError: %s", db.Name, db, err.Error())
		}
		for _, query := range db.Queries {
			params := make([]state.QueryParam, 0, len(query.Params))

			for _, p := range query.Params {
				params = append(params, state.QueryParam{
					Name:   p.Name,
					Format: p.Format,
				})
			}

			queryContext := state.QueryContext{
				Db:     &database,
				Name:   query.Name,
				Query:  query.Format,
				Params: params,
			}
			sqlQueryOptions = append(sqlQueryOptions, state.SelectableOption{
				Text:  fmt.Sprintf("%s: %s", db.Name, query.Name),
				Value: queryContext,
			})
		}
	}

	aStore = store.NewStore(state.State{
		Messages:           []state.MessageStruct{},
		SelectedMessageIdx: -1,
		Notification:       nil,
		AppActions:         []string{},
		ShowHeaders:        false,
		FocusedViews:       commons.NewStack(layout.DefaultView),
		Debug:              aConfiguration.Debug,
		InputMode:          false,
		SelectQueryPopup: state.SelectQueryPopupData{
			Text:        "Choose a query to run",
			Options:     sqlQueryOptions,
			SelectedIdx: 0,
		},
		FillQueryParamsPopup: state.FillQueryParamsPopupData{
			SelectedParamIdx: 0,
		},
	})

	aStore.AddReducer(func(s *state.State, a store.Action) {
		switch action := a.(type) {
		case state.NextMessage:
			if len(s.Messages) > s.SelectedMessageIdx+1 {
				s.SelectedMessageIdx++
			}
		case state.PrevMessage:
			if s.SelectedMessageIdx > 0 {
				s.SelectedMessageIdx--
			}
		case state.LoadMessages:
			if s.Messages != nil && len(s.Messages) > 0 {
				if err := rabbitmq.PublishMessagesToDlq(s.Messages, aConfiguration.Rabbitmq); err != nil {
					log.Printf("Failed to requeue messages, err: %s", err.Error())
				}
			}
			if messages, err := rabbitmq.LoadMessages(aConfiguration.Rabbitmq); err != nil {
				log.Printf("Failed to load messages, %s", err.Error())
			} else {
				s.Messages = messages
			}
		case state.RequeueMessages:
			if s.Messages != nil && len(s.Messages) > 0 {
				if err := rabbitmq.PublishMessagesToDlq(s.Messages, aConfiguration.Rabbitmq); err != nil {
					log.Printf("Failed to requeue messages, err: %s", err.Error())
				}
			}
		case state.ToggleShowHeaders:
			s.ShowHeaders = !s.ShowHeaders
		case state.DropMessage:
			s.Messages = append(s.Messages[:action.MessageIdx], s.Messages[action.MessageIdx+1:]...)
			if s.SelectedMessageIdx >= len(s.Messages) {
				s.SelectedMessageIdx--
			}
		case state.QueriesListNextOption:
			if s.SelectQueryPopup.SelectedIdx < len(s.SelectQueryPopup.Options)-1 {
				s.SelectQueryPopup.SelectedIdx += 1
			}
		case state.QueriesListPrevOption:
			if s.SelectQueryPopup.SelectedIdx > 0 {
				s.SelectQueryPopup.SelectedIdx -= 1
			}
		case state.ShowFillQueryParamsPopup:
			context, ok := s.SelectQueryPopup.Options[s.SelectQueryPopup.SelectedIdx].Value.(state.QueryContext)
			if !ok {
				log.Printf("Can't get query context from selected option value - invalid type")
				break
			}
			s.FillQueryParamsPopup = state.FillQueryParamsPopupData{
				Ctx:              context,
				SelectedParamIdx: 0,
			}
		case state.FillQueryParamsPopupNextField:
			if s.FillQueryParamsPopup.SelectedParamIdx < len(s.FillQueryParamsPopup.Ctx.Params)-1 {
				s.FillQueryParamsPopup.SelectedParamIdx++
			}
		case state.FillQueryParamsPopupPrevField:
			if s.FillQueryParamsPopup.SelectedParamIdx > 0 {
				s.FillQueryParamsPopup.SelectedParamIdx--
			}
		case state.StopInputMode:
			s.InputMode = false
		case state.StartInputMode:
			s.InputMode = true
		case state.RunSqlQuery:
			ctx := s.FillQueryParamsPopup.Ctx
			params := make(map[string]string)
			for _, p := range ctx.Params {
				params[p.Name] = fmt.Sprintf(p.Format, p.Value)
			}
			results := ctx.Db.Query(ctx.Query, params)

			s.DatabaseOutputs = &state.DatabaseData{
				Results: results,
			}

			sqlViewRows := layout.CalculateRows(results)

			s.SqlResultsView = &state.SqlResultsViewData{
				DX:    0,
				DY:    0,
				MaxDX: len(sqlViewRows[0]),
				MaxDY: len(sqlViewRows),
				Rows:  sqlViewRows,
			}
		case state.HideSqlResults:
			s.DatabaseOutputs = nil
		case state.SqlViewScrollDown:
			if s.SqlResultsView.DY < s.SqlResultsView.MaxDY {
				s.SqlResultsView.DY++
			}
		case state.SqlViewScrollUp:
			if s.SqlResultsView.DY > 0 {
				s.SqlResultsView.DY--
			}
		case state.SqlViewScrollRight:
			if s.SqlResultsView.DX < s.SqlResultsView.MaxDX {
				s.SqlResultsView.DX++
			}
		case state.SqlViewScrollLeft:
			if s.SqlResultsView.DX > 0 {
				s.SqlResultsView.DX--
			}
		}
	})

	appExit := make(chan bool, 1)
	appExitFunc := func() {
		appExit <- true
	}

	aLayout, err = layout.New(aStore, appExitFunc)
	if err != nil {
		log.Fatal(err)
	}

	go aLayout.Show()

	// Blocking until exit signal will be received
	<-appExit
}

func loadConfiguration() error {
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(configBytes, &aConfiguration)
	if err != nil {
		return err
	}

	return nil
}

func initLogger() {
	file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(io.MultiWriter(file, os.Stdout))
}
