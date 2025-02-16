package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"go.altair.com/todolist/pkg/structs"
)

func NewSqlStore(db *sqlx.DB) Store {
	return &sqlStore{
		db: db,
	}
}

type sqlStore struct {
	db *sqlx.DB
}

func (s *sqlStore) Update(action func(tx Txn) error) error {
	dbtx, err := s.db.Beginx()
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			_ = dbtx.Rollback()
			panic(r)
		}
	}()

	tx := &sqlStoreTxn{
		txn: dbtx,
	}

	err = action(tx)
	if err != nil {
		_ = dbtx.Rollback()
		return err
	}

	return dbtx.Commit()
}

type sqlStoreTxn struct {
	txn *sqlx.Tx
}

func readRecord(rows *sql.Rows, record *structs.TodoItem) error {
	return rows.Scan(
		&record.Id,
		&record.Item,
		&record.ItemOrder, 
	)
}


func (tx *sqlStoreTxn) DbTx() interface{} {
	return tx.txn
}

func (tx *sqlStoreTxn) Add(ctx context.Context, record *structs.TodoItem) error {
	_, err := tx.txn.ExecContext(ctx,
		tx.txn.Rebind("INSERT INTO TODOLIST(ID, ITEM, ITEM_ORDER) VALUES(?, ?,?)"),
		record.Id,
		record.Item, record.ItemOrder,
	)
	return err
}

func (tx *sqlStoreTxn) Delete(ctx context.Context, id string) error {
	result, err := tx.txn.ExecContext(ctx, tx.txn.Rebind("DELETE FROM TODOLIST WHERE ID=?"), id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("unknown id")
	}
	return nil
}

func (tx *sqlStoreTxn) Update(ctx context.Context, record *structs.TodoItem) error {
	result, err := tx.txn.ExecContext(ctx,
		tx.txn.Rebind(`UPDATE TODOLIST SET
			ITEM=?, ITEM_ORDER=?
			WHERE ID=?`),
		record.Item, record.ItemOrder,
		record.Id,
	)

	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("unknown id")
	}
	return nil
}

func (tx *sqlStoreTxn) Get(ctx context.Context, id string, item *structs.TodoItem) error {
	queryStmt := "SELECT ID, ITEM, ITEM_ORDER FROM TODOLIST WHERE ID=?"

	rows, err := tx.txn.QueryContext(ctx, tx.txn.Rebind(queryStmt), id)
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		return fmt.Errorf("unknown id")
	}

	if err := readRecord(rows, item); err != nil {
		return err
	}

	return nil
}

func (tx *sqlStoreTxn) List(ctx context.Context, items *structs.TodoItemList) error {
	queryStmt := "SELECT ID, ITEM, ITEM_ORDER FROM TODOLIST"

	rows, err := tx.txn.QueryContext(ctx, tx.txn.Rebind(queryStmt))
	if err != nil {
		return err
	}
	defer rows.Close()

	items.Items = make([]structs.TodoItem, 0)
	var record structs.TodoItem
	for rows.Next() {
		if err := readRecord(rows, &record); err != nil {
			return err
		}
		items.Items = append(items.Items, record)
		items.Count++
	}
	return nil
}
