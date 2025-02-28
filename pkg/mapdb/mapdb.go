package mapdb

import (
	"errors"
	pb "lyceum/pkg/api/test/api"
	"sync"
)

type DataMap struct {
	data map[string]*pb.Order
	mu   sync.RWMutex
}

func NewMap() *DataMap {
	return &DataMap{data: make(map[string]*pb.Order), mu: sync.RWMutex{}}
}

func (s *DataMap) Create(id string, item string, quantity int32) {
	itemObject := pb.Order{
		Item:     item,
		Quantity: quantity,
		Id:       id,
	}
	s.mu.Lock()
	s.data[id] = &itemObject
	s.mu.Unlock()
}

func (s *DataMap) Get(id string) (*pb.Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	itemObject, ok := s.data[id]
	if !ok {
		return &pb.Order{}, errors.New("item not found")
	}
	return itemObject, nil
}

func (s *DataMap) Update(id string, item string, quantity int32) (*pb.Order, error) {
	itemObject, err := s.Get(id)
	if err != nil {
		return &pb.Order{}, errors.New("item not found")
	}
	itemObject.Quantity = quantity
	itemObject.Item = item

	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[id] = itemObject

	return s.data[id], nil
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

func (s *DataMap) List() []*pb.Order {
	var array []*pb.Order
	s.mu.RLock()
	for _, itemObject := range s.data {
		array = append(array, itemObject)
	}
	s.mu.RUnlock()
	return array
}
