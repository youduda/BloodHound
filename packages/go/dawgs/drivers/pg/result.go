package pg

import (
	"github.com/jackc/pgx/v5"
	"github.com/specterops/bloodhound/dawgs/drivers"
	"github.com/specterops/bloodhound/dawgs/graph"
)

func NewValueMapper(values []any) *drivers.ValueMapper {
	return drivers.NewValueMapper(drivers.MapValue, values)
}

type queryError struct {
	err error
}

func (s queryError) Next() bool {
	return false
}

func (s queryError) Values() (graph.ValueMapper, error) {
	return nil, s.err
}

func (s queryError) Scan(targets ...any) error {
	return s.err
}

func (s queryError) Error() error {
	return s.err
}

func (s queryError) Close() {
}

type queryResult struct {
	rows pgx.Rows
}

func (s *queryResult) Next() bool {
	return s.rows.Next()
}

func (s *queryResult) Values() (graph.ValueMapper, error) {
	if values, err := s.rows.Values(); err != nil {
		return nil, err
	} else {
		return NewValueMapper(values), nil
	}
}

func (s *queryResult) Scan(targets ...any) error {
	return s.rows.Scan(targets...)
}

func (s *queryResult) Error() error {
	return s.rows.Err()
}

func (s *queryResult) Close() {
	s.rows.Close()
}
