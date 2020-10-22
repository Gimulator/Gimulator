package storage

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/Gimulator/protobuf/go/api"
	_ "github.com/mattn/go-sqlite3"
)

type Memory struct {
	db *sql.DB
}

var sqlite Memory

func init() {

	// Create db file
	err := createDBFile()
	if err != nil {
		log.Fatal(err)
	}

	// Open the sqlite file
	err = open(&sqlite)
	if err != nil {
		log.Fatal(err)
	}

	// Create table
	err = createTable(&sqlite)
	if err != nil {
		log.Fatal(err)
	}
}

func createDBFile() error {
	file, err := os.Create("sqlite-database.db")
	if err != nil {
		return err
	}
	file.Close()
	return nil
}

func open(m *Memory) error {
	db, err := sql.Open("sqlite3", "./sqlite-database.db")
	m.db = db
	if err != nil {
		return err
	}
	return nil
}

func createTable(m *Memory) error {
	createMessageTable := `CREATE TABLE message (
		"type" TEXT,
		"name" TEXT,
		"namespace" TEXT,
		"content" TEXT,
		"owner" TEXT,
		"creationtime" TEXT,
		PRIMARY KEY (type, name, namespace)
	  );`

	statement, err := m.db.Prepare(createMessageTable)
	if err != nil {
		return err
	}
	_, err = statement.Exec()
	if err != nil {
		return err
	}

	return nil
}

func GetDB() *Memory {
	return &sqlite
}

func (m *Memory) Put(message *api.Message) error {
	return m.put(message)
}

func (m *Memory) Delete(key *api.Key) error {
	return m.delete(key)
}

func (m *Memory) DeleteAll(key *api.Key) error {
	return m.deleteAll(key)
}

func (m *Memory) Get(key *api.Key) (*api.Message, error) {
	return m.get(key)
}

func (m *Memory) GetAll(key *api.Key) ([]*api.Message, error) {
	return m.getAll(key)
}

func (m *Memory) put(message *api.Message) error {
	err := m.validateKey(message.Key)
	if err != nil {
		return err
	}

	insertSQL := `INSERT INTO message VALUES (?, ?, ?, ?, ?, ?)`
	statement, err := m.db.Prepare(insertSQL)
	if err != nil {
		return err
	}
	_, err = statement.Exec(message.Key.Type, message.Key.Name, message.Key.Namespace, message.Content, message.Meta.Owner, message.Meta.CreationTime.String())
	if err != nil {
		return err
	}

	return nil
}

func (m *Memory) get(key *api.Key) (*api.Message, error) {
	err := m.validateKey(key)
	if err != nil {
		return nil, err
	}

	selectStatement := `SELECT * FROM message WHERE type = ? AND name = ? AND namespace = ?`
	row, err := m.db.Query(selectStatement, key.Type, key.Name, key.Namespace)
	if err != nil {
		return nil, err
	}
	defer row.Close()
	flag := false
	keyR := api.Key{}
	metaR := api.Meta{}
	messageR := api.Message{
		Content: "",
		Key:     &keyR,
		Meta:    &metaR,
	}
	for row.Next() {
		flag = true
		err = row.Scan(&messageR.Key.Type, &messageR.Key.Name, &messageR.Key.Namespace, &messageR.Content, &messageR.Meta.Owner, &messageR.Meta.CreationTime)
		if err != nil {
			return nil, err
		}
	}
	if !flag {
		return nil, fmt.Errorf("object with key=%v does not exist", *key)
	}

	return &messageR, nil
}

func (m *Memory) delete(key *api.Key) error {
	err := m.validateKey(key)
	if err != nil {
		return err
	}

	deleteStatement := `DELETE FROM message WHERE type = ? AND name = ? AND namespace = ? `
	statement, err := m.db.Prepare(deleteStatement)
	if err != nil {
		return err
	}
	_, err = statement.Exec(key.Type, key.Name, key.Namespace)
	if err != nil {
		return err
	}
	return nil
}

// TODO
func (m *Memory) deleteAll(key *api.Key) error {

	deleteStatement := `DELETE FROM message`
	statement, err := m.db.Prepare(deleteStatement)
	if err != nil {
		return err
	}
	_, err = statement.Exec()
	if err != nil {
		return err
	}
	return nil
}

// TODO
func (m *Memory) getAll(key *api.Key) ([]*api.Message, error) {
	return nil, nil
}

func (m *Memory) validateKey(key *api.Key) error {
	if key.Name == "" {
		return fmt.Errorf("invalid key with empty Name")
	}
	if key.Namespace == "" {
		return fmt.Errorf("invalid key with empty Namespace")
	}
	if key.Type == "" {
		return fmt.Errorf("invalid key with empty Type")
	}
	return nil
}
