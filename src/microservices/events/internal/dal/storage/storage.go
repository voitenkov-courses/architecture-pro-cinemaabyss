package storage

import "context"

type Storage struct{}

func New() *Storage {
	return &Storage{}
}

func (s *Storage) Connect() error {
	// Not implemented yet
	return nil
}

func (s *Storage) Close() error {
	// Not implemented yet
	return nil
}

func (s *Storage) SaveSensorUpdates(ctx context.Context) error {
	// Not implemented yet
	return nil
}
