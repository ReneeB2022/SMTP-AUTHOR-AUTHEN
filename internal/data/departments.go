package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ReneeB2022/test1/internal/validator"
)

// each name begins with uppercase so that they are exportable/public
type Department struct {
	Department_id   int64     `json:"id"`
	Department_name string    `json:"Deptname"` // the department data
	Department_code string    `json:"Deptcode"`
	CreatedAt       time.Time `json:"-"` // database timestamp
}

type DepartmentModel struct {
	DB *sql.DB
}

// Insert a new row in the departments table
// Expects a pointer to the actual department
func (c DepartmentModel) Insert(dept *Department) error {
	// the SQL query to be executed against the database table
	query := `
        INSERT INTO Departments (department_name, department_code)
        VALUES ($1, $2)
        RETURNING department_id, created_at
        `
	// the actual values to replace $1, and $2
	args := []any{dept.Department_name, dept.Department_code}

	// Create a context with a 3-second timeout. No database
	// operation should take more than 3 seconds or we will quit it
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// execute the query against the departments database table. We ask for the the
	// department_id and created_at to be sent back to us which we will use
	// to update the Department struct later on
	return c.DB.QueryRowContext(ctx, query, args...).Scan(
		&dept.Department_id,
		&dept.CreatedAt)
}

// Get a specific Department from the Departments table
func (c DepartmentModel) Get(id int64) (*Department, error) {
	// check if the id is valid
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	// the SQL query to be executed against the database table
	query := `
        SELECT department_id, created_at, department_name, department_code
        FROM Departments
        WHERE department_id = $1
      `
	// declare a variable of type Departments to store the returned Department
	var dept Department

	// Set a 3-second context/timer
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, id).Scan(
		&dept.Department_id,
		&dept.CreatedAt,
		&dept.Department_name,
		&dept.Department_code,
	)
	// check for which type of error
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &dept, nil
}

// Update a specific Department from the Departments table
func (c DepartmentModel) Update(dept *Department) error {
	// The SQL query to be executed against the database table
	// Every time we make an update
	query := `
        UPDATE Departments
        SET department_name = $1, department_code = $2
        WHERE department_id = $3
        RETURNING department_id
      `

	args := []any{dept.Department_name, dept.Department_code, dept.Department_id}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.DB.QueryRowContext(ctx, query, args...).Scan(&dept.Department_id)

}

// Delete a specific Department from the Departments table
func (c DepartmentModel) Delete(id int64) error {

	// check if the id is valid
	if id < 1 {
		return ErrRecordNotFound
	}
	// the SQL query to be executed against the database table
	query := `
        DELETE FROM Departments
        WHERE department_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// ExecContext does not return any rows unlike QueryRowContext.
	// It only returns  information about the the query execution
	// such as how many rows were affected
	result, err := c.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	// Were any rows  delete?
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	// Probably a wrong id was provided or the client is trying to
	// delete an already deleted Department
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil

}

// Get all Departments
func (c DepartmentModel) GetAll(deptname string, deptcode string, filters Filters) ([]*Department, Metadata, error) {

	query := fmt.Sprintf(`
        SELECT COUNT(*) OVER(),department_id, created_at, department_name, department_code
        FROM Departments
        WHERE (to_tsvector('simple', department_name) @@
              plainto_tsquery('simple', $1) OR $1 = '') 
        AND (to_tsvector('simple', department_code) @@ 
             plainto_tsquery('simple', $2) OR $2 = '') 
		ORDER BY %s %s, department_id ASC 
		LIMIT $3 OFFSET $4
     `, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// QueryContext returns multiple rows.
	rows, err := c.DB.QueryContext(ctx, query, deptname, deptcode, filters.limit(), filters.offset())
	totalRecords := 0

	// we will store the address of each Department in our slice
	depts := []*Department{}

	// process each row that is in rows

	for rows.Next() {
		var dept Department
		err := rows.Scan(
			&totalRecords,
			&dept.Department_id,
			&dept.CreatedAt,
			&dept.Department_name,
			&dept.Department_code,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		// add the row to our slice
		depts = append(depts, &dept)
	} // end of for loop
	// after we exit the loop we need to check if it generated any errors
	err = rows.Err()
	if err != nil {
		return nil, Metadata{}, err
	}
	// Create the metadata
	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)

	return depts, metadata, nil

}

// Create a function that performs the validation checks
func Validatedepartment(v *validator.Validator, dept *Department) {
	// check if the department name field is empty
	v.Check(dept.Department_name != "", "department name", "must be provided")
	// check if the department code is empty
	v.Check(dept.Department_code != "", "department code", "must be provided")
	// check if the department name field is empty
	v.Check(len(dept.Department_name) <= 100, "Department name", "must not be more than 100 bytes long")
	// check if the department code field is empty
	v.Check(len(dept.Department_code) <= 25, "Department code", "must not be more than 25 bytes long")
}
