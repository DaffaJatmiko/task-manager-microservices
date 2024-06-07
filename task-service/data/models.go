package data

import (
	"context"
	"database/sql"
	"log"
	"time"
)

const dbTimeout = time.Second * 3

var db *sql.DB

// New is the function used to create an instance of the data package. It returns the type
// Model, which embeds all the types we want to be available to our application.
func New(dbPool *sql.DB) Models {
	db = dbPool

	return Models{
		Task: Task{},
	}
}

// Models is the type for this package. Note that any model that is included as a member
// in this type is available to us throughout the application, anywhere that the
// app variable is used, provided that the model is also added in the New function.
type Models struct {
	Task Task
}

// Task is the structure which holds one task from the database.
type Task struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	UserID      int       `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GetAll returns a slice of all tasks, sorted by created_at
func (t *Task) GetAll() ([]*Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `select id, name, description, user_id, created_at, updated_at from tasks order by created_at`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task

	for rows.Next() {
		var task Task
		err := rows.Scan(
			&task.ID,
			&task.Name,
			&task.Description,
			&task.UserID,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			log.Println("Error scanning", err)
			return nil, err
		}

		tasks = append(tasks, &task)
	}

	return tasks, nil
}

// GetOne returns one task by id
func (t *Task) GetOne(id int) (*Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `select id, name, description, user_id, created_at, updated_at from tasks where id = ?`

	var task Task
	row := db.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&task.ID,
		&task.Name,
		&task.Description,
		&task.UserID,
		&task.CreatedAt,
		&task.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &task, nil
}

// GetTasksByUserID returns tasks by user ID
func (t *Task) GetTasksByUserID(userID int) ([]Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `select id, name, description, user_id, created_at, updated_at from tasks where user_id = ?`

	rows, err := db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task

	for rows.Next() {
		var task Task
		err := rows.Scan(
			&task.ID,
			&task.Name,
			&task.Description,
			&task.UserID,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			log.Println("Error scanning", err)
			return nil, err
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

// Insert inserts a new task into the database, and returns the ID of the newly inserted row
func (t *Task) Insert(task Task) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	log.Println("Inserting task", task)

	stmt := `insert into tasks (name, description, user_id, created_at, updated_at)
		values (?, ?, ?, ?, ?)`

	res, err := db.ExecContext(ctx, stmt,
		task.Name,
		task.Description,
		task.UserID,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		log.Println("Error inserting row", err)
		return 0, err
	}

	newID, err := res.LastInsertId()
	if err != nil {
		log.Println("Error getting last insert ID", err)
		return 0, err
	}

	return int(newID), nil

}

// Update updates one task in the database, using the information
// stored in the receiver t
func (t *Task) Update(task *Task) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	log.Println("Updating task", task)

	stmt := `update tasks set
	name = ?,
	description = ?,
	user_id = ?,
	updated_at = ?
	where id = ?`

	_, err := db.ExecContext(ctx, stmt,
			task.Name,
			task.Description,
			task.UserID,
			time.Now(),
			task.ID,  
	)

	if err != nil {
		log.Println("Error updating", err)
			return err
	}

	log.Println("Updated", t)
	return nil
}

// Delete deletes one task from the database, by Task.ID
func (t *Task) Delete(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `delete from tasks where id = ?`

	_, err := db.ExecContext(ctx, stmt, id)
	if err != nil {
		return err
	}

	return nil
}
