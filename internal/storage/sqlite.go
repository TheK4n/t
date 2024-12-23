
package storage

import (
	"database/sql"
	"fmt"
	"strconv"
	"io"
	_ "github.com/mattn/go-sqlite3"
)

type SqlTasksStorage struct {
	dbPath string
}

func (ts *SqlTasksStorage) GetNamespaces() ([]string, error) {
	db, err := sql.Open("sqlite3", ts.dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("select name from namespaces;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	namespaces := []string{}

	for rows.Next() {
		namespace := ""
		err := rows.Scan(&namespace)
		if err != nil {
			fmt.Println(err)
			continue
		}
		namespaces = append(namespaces, namespace)
	}

	return namespaces, nil
}

func (ts *SqlTasksStorage) Count(namespace string) (uint, error) {
	db, err := sql.Open("sqlite3", ts.dbPath)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	row := db.QueryRow("select COUNT(1) from tasks where namespace = $1;", namespace)

	namespacesCount := 0
	err = row.Scan(&namespacesCount)
	if err != nil {
		return 0, err
	}

	return uint(namespacesCount), nil
}

func (ts *SqlTasksStorage) GetSorted(namespace string) ([]string, error) {
	db, err := sql.Open("sqlite3", ts.dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("select COUNT(1) from tasks where namespace = '$1' sorted by born;", namespace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []string{}

	for rows.Next() {
		task := ""
		err := rows.Scan(&task)
		if err != nil {
			fmt.Println(err)
			continue
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (ts *SqlTasksStorage) Add(namespace string, name string) error {
	db, err := sql.Open("sqlite3", ts.dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("insert into tasks(name, namespace) values('$1', '$2');", name, namespace)

	if err != nil {
		return err
	}
	return nil
}

func (ts *SqlTasksStorage) GetNameByIndex(namespace string, index string) (string, error) {
	tasks, err := ts.GetSorted(namespace)
	if err != nil {
		return "", err
	}

	taskIndex, err := strconv.Atoi(index)
	if err != nil || taskIndex > len(tasks) || taskIndex < 1 {
		return "", fmt.Errorf("Wrong task index: %s", index)
	}

	return tasks[taskIndex-1], nil
}

func (ts *SqlTasksStorage) GetContentByName(namespace string, name string) ([]byte, error) {
	db, err := sql.Open("sqlite3", ts.dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	row := db.QueryRow("select content from tasks where namespace = $1 and name = $2;", namespace, name)

	taskContent := ""
	err = row.Scan(&taskContent)
	if err != nil {
		return nil, err
	}

	return []byte(taskContent), nil
}


func (ts *SqlTasksStorage) GetContentByIndex(namespace string, index string) ([]byte, error) {
	taskNameToRead, err := ts.GetNameByIndex(namespace, index)
	if err != nil {
		return nil, err
	}

	content, err := ts.GetContentByName(namespace, taskNameToRead)
	if err != nil {
		return nil, err
	}

	return content, nil
}



func (ts *SqlTasksStorage) WriteByName(namespace string, name string, r io.Reader) error {
	db, err := sql.Open("sqlite3", ts.dbPath)
	if err != nil {
		return err
	}
	defer db.Close()


	content := make([]byte, 0, 1024)
	_, err = r.Read(content)
	if err != nil {
		return err
	}

	_, err = db.Exec("update tasks set content = '$1' where name = $2;", string(content), name)
	if err != nil {
		return err
	}

	return nil
}

func (ts *SqlTasksStorage) WriteByIndex(namespace string, index string, r io.Reader) error {
	taskNameToWrite, err := ts.GetNameByIndex(namespace, index)

	db, err := sql.Open("sqlite3", ts.dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	content := make([]byte, 0, 1024)
	_, err = r.Read(content)
	if err != nil {
		return err
	}

	_, err = db.Exec("update tasks set content = '$1' where name = $2;", string(content), taskNameToWrite)
	if err != nil {
		return err
	}

	return nil
}

func (ts *SqlTasksStorage) CountLines(namespace string, name string) (uint, error) {
	content, err := ts.GetContentByName(namespace, name)
	if err != nil {
		return 0, err
	}

	return countRune(string(content), '\n'), nil
}

func countRune(s string, r rune) uint {
    var count uint = 0
    for _, c := range s {
        if c == r {
            count++
        }
    }
    return count
}
