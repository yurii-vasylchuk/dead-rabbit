package state

type NextMessage struct {
}

type PrevMessage struct {
}

type LoadMessages struct {
}

type FocusView struct {
	ViewName string
}

type FocusNextView struct {
}

type RequeueMessages struct {
}

type ToggleShowHeaders struct {
}

type ForceRedraw struct {
}

type DropMessage struct {
	MessageIdx int
}

type ShowQueriesListPopup struct {
}

type HidePopup struct {
}

type QueriesListNextOption struct {
}

type QueriesListPrevOption struct {
}

type ShowFillQueryParamsPopup struct {
}

type FillQueryParamsPopupNextField struct {
}

type FillQueryParamsPopupPrevField struct {
}

type StartInputMode struct {
}

type StopInputMode struct {
}

type Input struct {
	Ch rune
}

type InputBackspace struct {
}

type RunSqlQuery struct {
}

type HideSqlResults struct {
}

type SqlViewScrollDown struct {
}

type SqlViewScrollUp struct {
}

type SqlViewScrollLeft struct {
}

type SqlViewScrollRight struct {
}
