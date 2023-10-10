// go:build !inmemory

package repository

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/xor111xor/pomodoro-go/internal/models"
	"sync"
	"time"
)

const (
	createTableInterval string = `CREATE TABLE IF NOT EXISTS "interval" (
		"id" INTEGER,
		"start_time" DATETIME NOT NULL,
		"planned_duration" INTEGER DEFAULT 0,
		"actual_duration" INTEGER DEFAULT 0,
		"category" TEXT NOT NULL,
		"state" INTEGER DEFAULT 1,
		PRIMARY KEY("id")
		);`
)

type dbRepo struct {
	db *sql.DB
	sync.RWMutex
}

func NewSQLite3Repo(dbfile string) (*dbRepo, error) {
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	if _, err := db.Exec(createTableInterval); err != nil {
		return nil, err
	}
	return &dbRepo{
		db: db,
	}, nil
}

func (r *dbRepo) Create(i models.Interval) (int64, error) {
	// Create entry in the repository
	r.Lock()
	defer r.Unlock()

	// Prepare INSERT statements
	insStmt, err := r.db.Prepare("INSERT INTO interval VALUES(NULL, ?,?,?,?,?)")
	if err != nil {
		return 0, err
	}
	defer insStmt.Close()

	// Exec INSERT statements
	res, err := insStmt.Exec(i.TimeStart, i.TimePlanning,
		i.TimeActual, i.Category, i.State)
	if err != nil {
		return 0, err
	}

	var id int64
	if id, err = res.LastInsertId(); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *dbRepo) Update(i models.Interval) error {
	// Update entry in the repository
	r.Lock()
	defer r.Unlock()

	// Prepare UPDATE statements
	updStmt, err := r.db.Prepare(
		"UPDATE interval SET start_time=?, actual_duration=?, state=? WHERE id=?")
	if err != nil {
		return err
	}
	defer updStmt.Close()

	// Exec UPDATE statements
	res, err := updStmt.Exec(i.TimeStart, i.TimeActual, i.State, i.ID)
	if err != nil {
		return err
	}

	// UPDATE results
	_, err = res.RowsAffected()
	return err
}

func (r *dbRepo) ByID(id int64) (models.Interval, error) {
	// Search item in the repository by ID
	r.RLock()
	defer r.RUnlock()

	row := r.db.QueryRow("SELECT * FROM interval WHERE id=?", id)

	i := models.Interval{}
	err := row.Scan(&i.ID, &i.TimeStart, &i.TimePlanning,
		&i.TimeActual, &i.Category, &i.State)

	return i, err
}
func (r *dbRepo) Last() (models.Interval, error) {
	// Search last item in the repository
	r.RLock()
	defer r.RUnlock()

	i := models.Interval{}

	err := r.db.QueryRow("SELECT * FROM interval ORDER BY id desc LIMIT 1").Scan(
		&i.ID, &i.TimeStart, &i.TimePlanning,
		&i.TimeActual, &i.Category, &i.State,
	)

	if err == sql.ErrNoRows {
		return i, models.ErrNoIntervals
	}

	if err != nil {
		return i, err
	}

	return i, nil
}

func (r *dbRepo) Breaks(n int) ([]models.Interval, error) {
	// Return last breaks for count
	r.RLock()
	defer r.RUnlock()

	stmt := `SELECT * FROM interval WHERE category LIKE '%Break'
	ORDER BY id DESC LIMIT ?`

	rows, err := r.db.Query(stmt, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data := []models.Interval{}
	for rows.Next() {
		i := models.Interval{}
		err := rows.Scan(&i.ID, &i.TimeStart, &i.TimePlanning,
			&i.TimeActual, &i.Category, &i.State)

		if err != nil {
			return nil, err
		}

		data = append(data, i)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return data, nil

}

func (r *dbRepo) CategorySummary(day time.Time,
	filter string) (time.Duration, error) {
	//Return a daily summary
	r.RLock()
	defer r.RUnlock()

	stmt := `SELECT sum(actual_duration) FROM interval
	WHERE category LIKE ? AND
	strftime('%Y-%m-%d', start_time, 'localtime')=
	strftime('%Y-%m-%d', ?, 'localtime')`

	var ds sql.NullInt64
	err := r.db.QueryRow(stmt, filter, day).Scan(&ds)

	var d time.Duration
	if ds.Valid {
		d = time.Duration(ds.Int64)
	}
	return d, err
}
