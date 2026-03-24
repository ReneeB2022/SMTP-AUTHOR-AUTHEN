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
type Program struct {
	Program_id    int64     `json:"id"`
	Programs      string    `json:"program"`
	Program_code  string    `json:"Procode"`
	Department_id int64     `json:"-"`
	Faculty       string    `json:"department"`
	CreatedAt     time.Time `json:"-"` // database timestamp
}

type ProgramModel struct {
	DB *sql.DB
}

// Insert a new row in the Programs table
// Expects a pointer to the actual Program
func (c ProgramModel) Insert(pro *Program) error {
	// the SQL query to be executed against the database table
	query := `
        INSERT INTO Programs (program, program_code, department_id)
        VALUES ($1, $2, $3)
        RETURNING program_id, created_at
        `
	// the actual values to replace $1, and $2
	args := []any{pro.Programs, pro.Program_code, pro.Department_id}

	// Create a context with a 3-second timeout. No database
	// operation should take more than 3 seconds or we will quit it
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// execute the query against the Programs database table. We ask for the the
	// Program_id and created_at to be sent back to us which we will use
	// to update the Program struct later on
	return c.DB.QueryRowContext(ctx, query, args...).Scan(
		&pro.Program_id,
		&pro.CreatedAt)
}

// Get a specific Program from the Programs table
func (c ProgramModel) Get(id int64) (*Program, error) {
	// check if the id is valid
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	// the SQL query to be executed against the database table
	query := `
        SELECT program_id, Programs.created_at, program, program_code, Programs.department_id, department_code
        FROM Programs
		INNER JOIN Departments ON Programs.department_id = Departments.department_id
        WHERE program_id = $1
      `
	// declare a variable of type Programs to store the returned Program
	var pro Program

	// Set a 3-second context/timer
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, id).Scan(
		&pro.Program_id,
		&pro.CreatedAt,
		&pro.Programs,
		&pro.Program_code,
		&pro.Department_id,
		&pro.Faculty,
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
	return &pro, nil
}

// Update a specific Program from the Programs table
func (c ProgramModel) Update(pro *Program) error {
	// The SQL query to be executed against the database table
	// Every time we make an update
	query := `
        UPDATE Programs
        SET program = $1, program_code = $2, department_id = $3
        WHERE program_id = $4
        RETURNING program_id
      `

	args := []any{pro.Programs, pro.Program_code, pro.Department_id, pro.Program_id}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.DB.QueryRowContext(ctx, query, args...).Scan(&pro.Program_id)

}

// Delete a specific Program from the Programs table
func (c ProgramModel) Delete(id int64) error {

	// check if the id is valid
	if id < 1 {
		return ErrRecordNotFound
	}
	// the SQL query to be executed against the database table
	query := `
        DELETE FROM Programs
        WHERE program_id = $1`

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
	// delete an already deleted Program
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil

}

// Get all Programs
func (c ProgramModel) GetAll(procode string, deptid int64, filters Filters) ([]*Program, Metadata, error) {

	query := fmt.Sprintf(`
        SELECT COUNT(*) OVER(), program_id, Programs.created_at, program,  program_code, Programs.department_id, 
		department_code
        FROM Programs
		INNER JOIN Departments ON Programs.department_id = Departments.department_id
        WHERE (to_tsvector('simple', program_code) @@
              plainto_tsquery('simple', $1) OR $1 = '') 
        AND (Programs.department_id = $2 OR $2 = 0)
		ORDER BY %s %s, program_id ASC 
		LIMIT $3 OFFSET $4
     `, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// QueryContext returns multiple rows.
	rows, err := c.DB.QueryContext(ctx, query, procode, deptid, filters.limit(), filters.offset())
	totalRecords := 0

	// we will store the address of each Program in our slice
	pros := []*Program{}

	// process each row that is in rows

	for rows.Next() {
		var pro Program
		err := rows.Scan(
			&totalRecords,
			&pro.Program_id,
			&pro.CreatedAt,
			&pro.Programs,
			&pro.Program_code,
			&pro.Department_id,
			&pro.Faculty,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		// add the row to our slice
		pros = append(pros, &pro)
	} // end of for loop
	// after we exit the loop we need to check if it generated any errors
	err = rows.Err()
	if err != nil {
		return nil, Metadata{}, err
	}
	// Create the metadata
	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)

	return pros, metadata, nil

}

// Create a function that performs the validation checks
func Validateprograms(v *validator.Validator, pro *Program) {
	// check if the Program name field is empty
	v.Check(pro.Programs != "", " Full Program name", "must be provided")

	v.Check(pro.Program_code != "", "Program name", "must be provided")
	// check if the Program code is empty
	v.Check(pro.Department_id > 0, "Department id", "must be greater than zero")

	v.Check(pro.Department_id < 5, "Department id", "cannot be greater than 4")

	v.Check(pro.Department_id > 0, "Valid Department id", "must be provided")
	// check if the Program name field is empty
	v.Check(len(pro.Program_code) <= 100, "Program code", "must not be more than 100 bytes long")

	v.Check(len(pro.Programs) <= 100, "Program code", "must not be more than 100 bytes long")
	// check if the Program code field is empty
}
