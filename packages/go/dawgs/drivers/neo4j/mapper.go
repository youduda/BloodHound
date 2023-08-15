package neo4j

import (
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
	"github.com/specterops/bloodhound/dawgs/drivers"
	"github.com/specterops/bloodhound/dawgs/graph"
	"time"
)

func AsTime(value any) (time.Time, error) {
	switch typedValue := value.(type) {
	case dbtype.Time:
		return typedValue.Time(), nil

	case dbtype.LocalTime:
		return typedValue.Time(), nil

	case dbtype.Date:
		return typedValue.Time(), nil

	case dbtype.LocalDateTime:
		return typedValue.Time(), nil

	default:
		return drivers.AsTime(value)
	}
}

func MapValue(rawValue, target any) error {
	switch typedTarget := target.(type) {
	case *time.Time:
		if value, err := AsTime(rawValue); err != nil {
			return err
		} else {
			*typedTarget = value
		}

	case *dbtype.Relationship:
		if value, typeOK := rawValue.(dbtype.Relationship); !typeOK {
			return fmt.Errorf("unexecpted type %T will not negotiate to *dbtype.Relationship", rawValue)
		} else {
			*typedTarget = value
		}

	case *graph.Relationship:
		if value, typeOK := rawValue.(dbtype.Relationship); !typeOK {
			return fmt.Errorf("unexecpted type %T will not negotiate to *dbtype.Relationship", rawValue)
		} else {
			*typedTarget = *newRelationship(value)
		}

	case *dbtype.Node:
		if value, typeOK := rawValue.(dbtype.Node); !typeOK {
			return fmt.Errorf("unexecpted type %T will not negotiate to *dbtype.Node", rawValue)
		} else {
			*typedTarget = value
		}

	case *graph.Node:
		if value, typeOK := rawValue.(dbtype.Node); !typeOK {
			return fmt.Errorf("unexecpted type %T will not negotiate to *dbtype.Node", rawValue)
		} else {
			*typedTarget = *newNode(value)
		}

	case *graph.Path:
		if value, typeOK := rawValue.(dbtype.Path); !typeOK {
			return fmt.Errorf("unexecpted type %T will not negotiate to *dbtype.Path", rawValue)
		} else {
			*typedTarget = newPath(value)
		}

	default:
		return drivers.MapValue(rawValue, target)
	}

	return nil
}

func NewValueMapper(values []any) *drivers.ValueMapper {
	return drivers.NewValueMapper(MapValue, values)
}
