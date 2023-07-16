package maya

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/edervzz/maya/cons"
	"github.com/edervzz/maya/internal/fcat"
	"github.com/edervzz/maya/internal/hcons"
	"github.com/edervzz/maya/internal/sqlb"

	"time"

	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

type dbContext struct {
	dbClient         *sql.DB
	trx              *sql.Tx
	logger           *zap.Logger
	connectionString string
	dbname           string
	migrations       []Migration
	enqueueTable     []enqueue
	isMigratedDone   bool
}

type ConnectionString struct {
	Server string
	DbName string
	Port   string
	User   string
	Pass   string
}

type enqueue struct {
	id         string
	entityName string
}

// constructor
func NewDbContext(conn ConnectionString, logger *zap.Logger, migrations []Migration) IDbContext {
	sort.Sort(ByID(migrations))
	return &dbContext{
		dbClient:         nil,
		trx:              nil,
		connectionString: fmt.Sprintf("%v:%v@tcp(%v:%v)/", conn.User, conn.Pass, conn.Server, conn.Port),
		logger:           logger,
		dbname:           conn.DbName,
		migrations:       migrations,
		enqueueTable:     []enqueue{},
	}
}

func (h *dbContext) Enqueue(ctx context.Context, id string, entityPtr any, info string) error {
	if err := h.healthCheck(); err != nil {
		return err
	}

	if reflect.TypeOf(entityPtr).Kind() != reflect.Pointer {
		return errors.New("entity param must be a pointer")
	}

	tableName := fcat.EnrichTableName(entityPtr)
	if tableName == "" {
		h.logger.Info("entity not found or misformmed")
		return errors.New("entity not found or misformmed")
	}

	if h.trx != nil {
		h.logger.Info("enqueue cannot be called in middle of transactions")
		return errors.New("enqueue cannot be called in middle of transactions")
	}
	datetime := time.Now().UTC().Format(time.DateTime)
	sqlresult, err := h.dbClient.Exec(
		"INSERT INTO _MayaLocks (id, entity, created_at, info) VALUES (?,?,?,?);",
		id,
		tableName,
		datetime,
		info,
	)
	if err != nil {
		h.logger.Info(err.Error())
		return err
	}
	if sqlresult == nil {
		h.logger.Info("enqueue cannot be reached, no result")
		return errors.New("enqueue cannot be reached, no result")
	}
	if lastid, _ := sqlresult.LastInsertId(); lastid < 0 {
		h.logger.Info("enqueue cannot be reached, no lastId")
		return errors.New("enqueue cannot be reached, no lastId")
	}

	h.enqueueTable = append(h.enqueueTable, enqueue{
		id:         id,
		entityName: tableName,
	})

	return nil
}

func (h *dbContext) Dequeue(ctx context.Context) error {

	if len(h.enqueueTable) == 0 {
		return nil
	}

	if h.trx != nil {
		h.logger.Info("dequeue cannot be called in middle of transactions")
		return errors.New("dequeue cannot be called in middle of transactions")
	}
	errlog := []string{}
	for _, e := range h.enqueueTable {
		_, err := h.dbClient.Exec(
			"DELETE FROM _MayaLocks WHERE id = ? AND entity = ?;",
			e.id,
			e.entityName,
		)
		if err != nil {
			errlog = append(errlog, err.Error())
			h.logger.Info(err.Error())
		}
	}

	if len(errlog) != 0 {
		return errors.New(strings.Join(errlog, " : "))
	}

	return nil
}

// begin a new transaction
func (h *dbContext) BeginTransaction(ctx context.Context) error {
	if err := h.healthCheck(); err != nil {
		return err
	}
	var err error = nil
	options := sql.TxOptions{Isolation: sql.LevelDefault, ReadOnly: false}
	if h.trx, err = h.dbClient.BeginTx(ctx, &options); err != nil {
		h.logger.Info(err.Error())
	}
	return err
}

// commit transaction
func (h *dbContext) Commit() error {
	var err error = nil
	if h.trx != nil {
		err = h.trx.Commit()
	}
	h.trx = nil
	return err
}

// rollback transaction
func (h *dbContext) Rollback() error {
	var err error
	if h.trx != nil {
		err = h.trx.Rollback()
	}
	h.trx = nil
	return err
}

// add new entity
func (h dbContext) Add(ctx context.Context, entityPtr any) (sql.Result, error) {
	if reflect.TypeOf(entityPtr).Kind() != reflect.Pointer {
		return nil, errors.New("entity param must be a pointer")
	}
	// 1. prepare query and arguments
	var e interface{} = entityPtr
	_, isAuditable := e.(IAuditable)
	query, args, tabName := sqlb.BuildInsert(ctx, entityPtr, isAuditable)
	// 2. execute query
	sqlResult, err := h.trx.ExecContext(ctx, query, args...)

	if err != nil {
		err = errors.New(fmt.Sprintf("failed INSERT '%v': %v", tabName, err.Error()))
	}

	return sqlResult, err
}

// update entity
func (h dbContext) Update(ctx context.Context, entityPtr any) (sql.Result, error) {
	if reflect.TypeOf(entityPtr).Kind() != reflect.Pointer {
		return nil, errors.New("entity param must be a pointer")
	}
	// 1. prepare query and arguments
	var e interface{} = entityPtr
	_, isAuditable := e.(IAuditable)
	query, args, tabName := sqlb.BuildUpdate(ctx, entityPtr, isAuditable)
	// 2. execute query
	sqlResult, err := h.trx.ExecContext(ctx, query, args...)

	if err != nil {
		err = errors.New(fmt.Sprintln("failed UPDATE '&v': &v", tabName, err.Error()))
	}
	return sqlResult, err
}

// Sql
func (h dbContext) Sql(ctx context.Context, query string, args ...any) error {
	// 1. execute query
	_, err := h.trx.ExecContext(ctx, query, args...)

	if err != nil {
		err = errors.New(fmt.Sprintln("failed Sql: ", err.Error()))
	}
	return err
}

// Read entity
func (h dbContext) Read(ctx context.Context, entityPtr any, filter map[string]any) (*sql.Rows, error) {
	if err := h.healthCheck(); err != nil {
		return nil, err
	}
	query, values, tableName := sqlb.BuildRead(ctx, entityPtr, filter)
	rows, err := h.dbClient.QueryContext(ctx, query, values...)

	if err != nil {
		err = errors.New(fmt.Sprintln("failed READING '&v': &v", tableName, err.Error()))
	}
	return rows, err
}

func (h *dbContext) migrate() error {
	if h.isMigratedDone {
		return nil
	}

	// 1. check if db exists
	rows, err := h.dbClient.Query(
		fmt.Sprintf("SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = '%s'", h.dbname))
	if err != nil {
		return err
	}

	// migration history
	migrationsHistory := map[string]string{}

	if rows.Next() {
		// 2. use db
		if _, err = h.dbClient.Exec("USE " + h.dbname); err != nil {
			return err
		}
		h.logger.Info("db in use")
		// 3. read all migrations done
		if rows, err := h.dbClient.Query("SELECT id, version FROM _MayaMigrationsHistory"); err == nil {
			// collect migration done
			for rows.Next() {
				id, ver := "", ""
				rows.Scan(&id, &ver)
				migrationsHistory[id] = ver
			}
		} else {
			// create table lock
			if _, err = h.dbClient.Exec(hcons.LockTable); err != nil {
				return err
			}
			h.logger.Info("lock table created")
			// create migration table whether db exists but not table
			if _, err = h.dbClient.Exec(hcons.MigrationsTable); err != nil {
				return err
			}
			h.logger.Info("migration table created")
		}
	} else {
		// 4. create database from scratch
		if _, err = h.dbClient.Exec("CREATE DATABASE " + h.dbname); err != nil {
			return err
		}
		// 4.1 use database
		if _, err = h.dbClient.Exec("USE " + h.dbname); err != nil {
			return err
		}
		h.logger.Info("db created and in use")
		// 4.2 create table lock
		if _, err = h.dbClient.Exec(hcons.LockTable); err != nil {
			return err
		}
		h.logger.Info("lock table created")
		// 4.3 create migration table and run migration
		if _, err = h.dbClient.Exec(hcons.MigrationsTable); err != nil {
			return err
		}
		h.logger.Info("migration table created")
	}
	// migrate for any one did not migrate
	return h.runMigration(migrationsHistory)
}

func (h *dbContext) runMigration(migrationHistory map[string]string) error {
	version := "1.0"
	downMigration := false
	for _, t := range h.migrations {
		// discard migrations done
		if migrationHistory[t.ID] != "" {
			continue
		}
		// run migration
		for _, dd := range t.Up {
			_, err := h.dbClient.Exec(dd)
			if err != nil {
				h.logger.Info(err.Error())
				downMigration = true
				break
			}
		}
		if downMigration {
			for _, dd := range t.Down {
				_, err := h.dbClient.Exec(dd)
				if err != nil {
					return err
				}
			}
			break
		}
		h.dbClient.Exec("INSERT into _MayaMigrationsHistory (id, version) VALUES (?,?)", t.ID, version)
	}
	h.isMigratedDone = true
	return nil
}

// Connection health check. Try to open connection to DB and process migrations when they are pending and active.
func (h *dbContext) healthCheck() error {
	// 1. try open connection
	if h.dbClient == nil {
		if client, err := sql.Open("mysql", h.connectionString); err != nil {
			h.logger.Info(err.Error())
			os.Exit(1)
		} else {
			h.dbClient = client
		}
	}
	// 2. use db or migrate, migration is activated by default
	if isMigrate := os.Getenv(cons.DB_MIGRATE); strings.ToLower(isMigrate) == "false" {
		if _, err := h.dbClient.Exec("USE " + h.dbname); err != nil {
			return err
		}
	} else {
		return h.migrate()
	}
	return nil
}
