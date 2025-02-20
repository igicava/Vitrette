package mapdb

import (
	"errors"
	"lyceum/model"
	"sync"
)

type DataMap struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

func NewMap() *DataMap {
	return &DataMap{data: make(map[string]interface{}), mu: sync.RWMutex{}}
}

func (s *DataMap) Create(id string, item string, quantity int32) {
	itemObject := model.OrderStruct{
		Item:     item,
		Quantity: quantity,
		ID:       id,
	}
	s.mu.Lock()
	s.data[id] = itemObject
	s.mu.Unlock()
}

func (s *DataMap) Get(id string) (model.OrderStruct, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	itemObject, ok := s.data[id]
	if !ok {
		return model.OrderStruct{}, errors.New("item not found")
	}
	return itemObject.(model.OrderStruct), nil
}

func (s *DataMap) Update(id string, item string, quantity int32) (model.OrderStruct, error) {
	itemObject, err := s.Get(id)
	if err != nil {
		return model.OrderStruct{}, errors.New("item not found")
	}
	itemObject.Quantity = quantity
	itemObject.Item = item

	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[id] = itemObject

	return s.data[id].(model.OrderStruct), nil
}

func (s *DataMap) Delete(id string) error {
	_, err := s.Get(id)
	if err != nil {
		return errors.New("item not found")
	}
	s.mu.Lock()
	delete(s.data, id)
	s.mu.Unlock()
	return nil
}

func (s *DataMap) List() []model.OrderStruct {
	var array []model.OrderStruct
	s.mu.RLock()
	for _, itemObject := range s.data {
		array = append(array, itemObject.(model.OrderStruct))
	}
	s.mu.RUnlock()
	return array
}
