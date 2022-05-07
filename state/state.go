package state

import (
	"time"

	"DeadRabbit/commons"
)

type NotificationStruct struct {
	Value string
	At    time.Time
}

type FillQueryParamsPopupData struct {
	Ctx              QueryContext
	SelectedParamIdx int
}

type DatabaseData struct {
	Results QueryResults
}

type SqlResultsViewData struct {
	DX    int
	DY    int
	MaxDX int
	MaxDY int
	Rows  []string
}

type State struct {
	InputMode            bool
	Debug                bool
	Messages             []MessageStruct
	SelectedMessageIdx   int
	Notification         *NotificationStruct
	AppActions           []string
	ShowHeaders          bool
	FocusedViews         *commons.Stack[string]
	SelectQueryPopup     SelectQueryPopupData
	FillQueryParamsPopup FillQueryParamsPopupData
	DatabaseOutputs      *DatabaseData
	SqlResultsView       *SqlResultsViewData
}

type SelectQueryPopupData struct {
	Text        string
	Options     []SelectableOption
	SelectedIdx int
}

type MessageStruct struct {
	Body    string
	Headers map[string]any
}

type SelectableOption struct {
	Text  string
	Value any
}

type QueryParam struct {
	Name   string
	Value  string
	Format string
}

type QueryContext struct {
	Db     Repository
	Name   string
	Query  string
	Params []QueryParam
}

type QueryResults interface {
	GetHeaders() []string
	GetResults() []map[string]string
}

type Repository interface {
	Query(sql string, params map[string]string) QueryResults
}
