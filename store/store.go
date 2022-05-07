package store

import (
	"log"
	"reflect"

	"DeadRabbit/commons"
)

type Reducer[STATE any] func(s *STATE, action Action)
type Action interface {
}

type Store[STATE any] struct {
	aState           STATE
	subscribers      map[int]chan<- STATE
	subscriptionsIdx int
	reducers         map[int]Reducer[STATE]
	reducersIdx      int
}

func NewStore[STATE any](initial STATE) *Store[STATE] {
	return &(Store[STATE]{
		aState:           initial,
		subscriptionsIdx: 0,
		subscribers:      make(map[int]chan<- STATE),
		reducers:         make(map[int]Reducer[STATE]),
		reducersIdx:      0,
	})
}

func (s *Store[STATE]) Subscribe() (updates <-chan STATE, unsubscribe func()) {
	s.subscriptionsIdx++
	idx := s.subscriptionsIdx
	log.Printf("Adding new subscriber; ID #%d", idx)
	channel := make(chan STATE, 10)
	s.subscribers[idx] = channel
	return channel, func() {
		log.Printf("Subscriber #%d unsubscribing", idx)
		delete(s.subscribers, idx)
	}
}

func (s *Store[STATE]) AddReducer(r Reducer[STATE]) func() {
	s.reducersIdx++
	idx := s.reducersIdx
	log.Printf("Adding reducer #%d, [%s]", idx, commons.GetFunctionDescription(r))
	s.reducers[idx] = r

	return func() {
		log.Printf("Deleting reducer #%d", idx)
		delete(s.reducers, idx)
	}
}

func (s *Store[STATE]) Dispatch(action Action) {
	log.Printf("Received action %+v", action)
	for i, reducer := range s.reducers {
		reducer(&s.aState, action)
		log.Printf("State processed by reducer #%d", i)
	}

	for _, subscriber := range s.subscribers {
		subscriber <- s.aState
	}
	log.Printf("Action %T handling finished", action)
}

func (s Store[STATE]) GetCurrent() STATE {
	return s.aState
}

func isDebug(state any) bool {
	const debugFieldName = "Debug"
	debugValue := reflect.ValueOf(state).FieldByName(debugFieldName)
	return debugValue.IsValid() && debugValue.CanConvert(reflect.TypeOf(true)) && debugValue.Bool()
}
