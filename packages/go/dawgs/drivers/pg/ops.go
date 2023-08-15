package pg

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/specterops/bloodhound/dawgs/graph"
	"strings"
)

func GetKindID(tx graph.Transaction, kind graph.Kind) (int16, error) {
	var kindID int16

	result := tx.Run(`select * from kind where name = @name`, map[string]any{
		"name": kind.String(),
	})
	defer result.Close()

	if !result.Next() {
		return -1, pgx.ErrNoRows
	}

	return kindID, result.Scan(&kindID)
}

func InitSchemaUp(ctx context.Context, db graph.Database) error {
	if driver, typeOK := db.(*driver); !typeOK {
		return fmt.Errorf("graph database is not a PostgreSQL database")
	} else {
		for _, stmt := range strings.Split(graphSchemaSQLUp, ";") {
			if err := driver.Run(ctx, stmt, nil); err != nil {
				return err
			}
		}
	}

	return nil
}

func InitSchemaDown(ctx context.Context, db graph.Database) error {
	if driver, typeOK := db.(*driver); !typeOK {
		return fmt.Errorf("graph database is not a PostgreSQL database")
	} else {
		for _, stmt := range strings.Split(graphSchemaSQLDown, ";") {
			if err := driver.Run(ctx, stmt, nil); err != nil {
				return err
			}
		}
	}

	return nil
}
