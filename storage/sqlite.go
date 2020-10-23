package storage

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/Gimulator/protobuf/go/api"
	"github.com/golang/protobuf/ptypes/timestamp"
	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type SqliteMemory struct {
	db *sql.DB
}

var sqlite SqliteMemory

func init() {

	path := ":memory:"
	if err := sqlite.prepare(path); err != nil {
		log.Fatal(err)
	}

	// Create table
	err := sqlite.createTable()
	if err != nil {
		log.Fatal(err)
	}
}

func GetDB() *SqliteMemory {
	return &sqlite
}

func (m *SqliteMemory) Put(message *api.Message) error {
	return m.put(message)
}

func (m *SqliteMemory) Delete(key *api.Key) error {
	return m.delete(key)
}

func (m *SqliteMemory) DeleteAll(key *api.Key) error {
	return m.deleteAll(key)
}

func (m *SqliteMemory) Get(key *api.Key) (*api.Message, error) {
	return m.get(key)
}

func (m *SqliteMemory) GetAll(key *api.Key) ([]*api.Message, error) {
	return m.getAll(key)
}

func (m *SqliteMemory) createDBFile(path string) error {
	if path == ":memory:" {
		return nil
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	file.Close()
	return nil
}

func (m *SqliteMemory) open(path string) error {
	db, err := sql.Open("sqlite3", path)
	m.db = db
	if err != nil {
		return err
	}
	return nil
}

func (m *SqliteMemory) prepare(path string) error {
	// Create db file
	err := sqlite.createDBFile(path)
	if err != nil {
		return err
	}

	// Open the sqlite file
	err = sqlite.open(path)
	if err != nil {
		return err
	}

	return nil
}

func (m *SqliteMemory) createTable() error {
	createMessageTable := `CREATE TABLE message (
		"type" TEXT,
		"name" TEXT,
		"namespace" TEXT,
		"content" TEXT,
		"owner" TEXT,
		"creationtimeseconds" INTEGER,
		"creationtimenanos" INTEGER,
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

func (m *SqliteMemory) put(message *api.Message) error {
	err := m.validateMessage(message)
	if err != nil {
		return err
	}

	insertSQL := `INSERT INTO message VALUES (?, ?, ?, ?, ?, ?, ?)`
	statement, err := m.db.Prepare(insertSQL)
	if err != nil {
		return err
	}
	seconds, nanos, err := m.unmarshalTimestamp(message.Meta.CreationTime.String())
	if err != nil {
		return err
	}
	_, err = statement.Exec(message.Key.Type, message.Key.Name, message.Key.Namespace, message.Content, message.Meta.Owner, seconds, nanos)
	if err != nil {
		return err
	}

	return nil
}

func (m *SqliteMemory) get(key *api.Key) (*api.Message, error) {
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
		var seconds, nanos int
		err = row.Scan(&messageR.Key.Type, &messageR.Key.Name, &messageR.Key.Namespace, &messageR.Content, &messageR.Meta.Owner, &seconds, &nanos)
		if err != nil {
			return nil, err
		}
		messageR.Meta.CreationTime = m.marshalTimestamp(seconds, nanos)
	}
	if !flag {
		return nil, fmt.Errorf("object with key=%v does not exist", *key)
	}

	return &messageR, nil
}

func (m *SqliteMemory) delete(key *api.Key) error {
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

func (m *SqliteMemory) deleteAll(key *api.Key) error {

	deleteStatement := `DELETE FROM message WHERE type LIKE ? AND name LIKE ? AND namespace LIKE ?`
	statement, err := m.db.Prepare(deleteStatement)
	if err != nil {
		return err
	}
	var t, n, ns string
	if key.Type == "" {
		t = "%"
	} else {
		t = key.Type
	}
	if key.Name == "" {
		n = "%"
	} else {
		n = key.Name
	}
	if key.Namespace == "" {
		ns = "%"
	} else {
		ns = key.Namespace
	}
	_, err = statement.Exec(t, n, ns)
	if err != nil {
		return err
	}
	return nil
}

func (m *SqliteMemory) getAll(key *api.Key) ([]*api.Message, error) {
	selectStatement := `SELECT * FROM message WHERE type LIKE ? AND name LIKE ? AND namespace LIKE ?`
	var t, n, ns string
	if key.Type == "" {
		t = "%"
	} else {
		t = key.Type
	}
	if key.Name == "" {
		n = "%"
	} else {
		n = key.Name
	}
	if key.Namespace == "" {
		ns = "%"
	} else {
		ns = key.Namespace
	}
	row, err := m.db.Query(selectStatement, t, n, ns)
	if err != nil {
		return nil, err
	}
	defer row.Close()
	var messages []*api.Message
	for row.Next() {
		keyR := api.Key{}
		metaR := api.Meta{}
		messageR := api.Message{
			Content: "",
			Key:     &keyR,
			Meta:    &metaR,
		}
		var seconds, nanos int
		err = row.Scan(&messageR.Key.Type, &messageR.Key.Name, &messageR.Key.Namespace, &messageR.Content, &messageR.Meta.Owner, &seconds, &nanos)
		if err != nil {
			return nil, err
		}
		messageR.Meta.CreationTime = m.marshalTimestamp(seconds, nanos)
		messages = append(messages, &messageR)
	}
	if len(messages) == 0 {
		return nil, fmt.Errorf("object with key=%v does not exist", *key)
	}

	return messages, nil
}

func (m *SqliteMemory) validateKey(key *api.Key) error {
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

func (m *SqliteMemory) validateMessage(message *api.Message) error {
	if message.Key == nil {
		return fmt.Errorf("Invalid message with empty key")
	}
	if err := m.validateKey(message.Key); err != nil {
		return err
	}
	return nil
}

func (m *SqliteMemory) unmarshalTimestamp(t string) (int, int, error) {
	str := strings.Split(t, " ")
	if len(str) != 2 {
		str = strings.Split(t, "  ")
	}
	seconds, err := strconv.Atoi(strings.Split(str[0], ":")[1])
	if err != nil {
		return -1, -1, err
	}
	nanos, err := strconv.Atoi(strings.Split(str[1], ":")[1])
	if err != nil {
		return -1, -1, err
	}
	return seconds, nanos, nil
}

func (m *SqliteMemory) marshalTimestamp(seconds, nanos int) *timestamp.Timestamp {
	finalT := timestamppb.Timestamp{
		Seconds: int64(seconds),
		Nanos:   int32(nanos),
	}
	return &finalT
}
