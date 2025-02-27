package database

import (
	"fmt"

	"github.com/jackc/pgx"

	"cyndi-operator/controllers/config"
	. "cyndi-operator/controllers/config"
)

type BaseDatabase struct {
	Config     *DBParams
	connection *pgx.Conn
}

const connectionStringTemplate = "host=%s user=%s password=%s dbname=%s port=%s"

func NewBaseDatabase(config *config.DBParams) Database {
	return &BaseDatabase{
		Config: config,
	}
}

func (db *BaseDatabase) Connect() (err error) {
	if db.connection, err = GetConnection(db.Config); err != nil {
		return fmt.Errorf("Error connecting to %s:%s/%s as %s : %s", db.Config.Host, db.Config.Port, db.Config.Name, db.Config.User, err)
	}

	return nil
}

func (db *BaseDatabase) Close() error {
	if db.connection != nil {
		return db.connection.Close()
	}

	return nil
}

func (db *BaseDatabase) RunQuery(query string) (*pgx.Rows, error) {
	rows, err := db.connection.Query(query)

	if err != nil {
		return nil, fmt.Errorf("Error executing query %s, %w", query, err)
	}

	return rows, nil
}

func (db *BaseDatabase) Exec(query string) (result pgx.CommandTag, err error) {
	result, err = db.connection.Exec(query)

	if err != nil {
		return result, fmt.Errorf("Error executing query %s, %w", query, err)
	}

	return result, nil
}

func (db *BaseDatabase) getWhereClause(insightsOnly bool) string {
	if insightsOnly {
		return "WHERE canonical_facts ? 'insights_id'"
	}

	return ""
}

func (db *BaseDatabase) hostCountQuery(table string, insightsOnly bool) string {
	return fmt.Sprintf(`SELECT count(*) FROM %s %s`, table, db.getWhereClause(insightsOnly))
}

func (db *BaseDatabase) CountHosts(table string, insightsOnly bool) (int64, error) {
	// TODO: add modified_on filter
	// waiting on https://issues.redhat.com/browse/RHCLOUD-9545
	rows, err := db.RunQuery(db.hostCountQuery(table, insightsOnly))

	if err != nil {
		return -1, err
	}

	defer rows.Close()

	var response int64
	for rows.Next() {
		var count int64
		err = rows.Scan(&count)
		if err != nil {
			return -1, err
		}
		response = count
	}

	if err != nil {
		return -1, err
	}

	return response, err
}

func (db *BaseDatabase) hostIdQuery(table string, insightsOnly bool) string {
	return fmt.Sprintf(`SELECT id FROM %s %s ORDER BY id`, table, db.getWhereClause(insightsOnly))
}

func (db *BaseDatabase) GetHostIds(table string, insightsOnly bool) ([]string, error) {
	// TODO: add modified_on filter
	// waiting on https://issues.redhat.com/browse/RHCLOUD-9545
	rows, err := db.RunQuery(db.hostIdQuery(table, insightsOnly))

	var ids []string

	if err != nil {
		return ids, err
	}

	defer rows.Close()

	for rows.Next() {
		var id string
		err = rows.Scan(&id)

		if err != nil {
			return ids, err
		}

		ids = append(ids, id)
	}

	return ids, nil
}

func GetConnection(params *DBParams) (connection *pgx.Conn, err error) {
	connStr := fmt.Sprintf(
		connectionStringTemplate,
		params.Host,
		params.User,
		params.Password,
		params.Name,
		params.Port)

	if config, err := pgx.ParseDSN(connStr); err != nil {
		return nil, err
	} else {
		if connection, err = pgx.Connect(config); err != nil {
			return nil, err
		} else {
			return connection, nil
		}
	}
}
