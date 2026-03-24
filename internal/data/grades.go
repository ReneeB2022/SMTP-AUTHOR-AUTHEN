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
type Grade struct {
	Grade_id    int64     `json:"id"`
	Grades      string    `json:"Letter"`
	Grade_value float64   `json:"Value"`
	CreatedAt   time.Time `json:"-"` // database timestamp
}

type GradeModel struct {
	DB *sql.DB
}

// Insert a new row in the Grades table
// Expects a pointer to the actual Grade
func (c GradeModel) Insert(grade *Grade) error {
	// the SQL query to be executed against the database table
	query := `
        INSERT INTO Grades (letter_grade, grade_value)
        VALUES ($1, $2)
        RETURNING grade_id, created_at
        `
	// the actual values to replace $1, and $2
	args := []any{grade.Grades, grade.Grade_value}

	// Create a context with a 3-second timeout. No database
	// operation should take more than 3 seconds or we will quit it
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// execute the query against the Grades database table. We ask for the the
	// Grade_id and created_at to be sent back to us which we will use
	// to update the Grade struct later on
	return c.DB.QueryRowContext(ctx, query, args...).Scan(
		&grade.Grade_id,
		&grade.CreatedAt)
}

// Get a specific Grade from the Grades table
func (c GradeModel) Get(id int64) (*Grade, error) {
	// check if the id is valid
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	// the SQL query to be executed against the database table
	query := `
        SELECT grade_id, created_at, letter_grade, grade_value
        FROM Grades
        WHERE grade_id = $1
      `
	// declare a variable of type Grades to store the returned Grade
	var grade Grade

	// Set a 3-second context/timer
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, id).Scan(
		&grade.Grade_id,
		&grade.CreatedAt,
		&grade.Grades,
		&grade.Grade_value,
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
	return &grade, nil
}

// Update a specific Grade from the Grades table
func (c GradeModel) Update(grade *Grade) error {
	// The SQL query to be executed against the database table
	// Every time we make an update
	query := `
        UPDATE Grades
        SET letter_grade = $1, grade_value = $2
        WHERE grade_id = $3
        RETURNING grade_id
      `

	args := []any{grade.Grades, grade.Grade_value}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.DB.QueryRowContext(ctx, query, args...).Scan(&grade.Grade_id)

}

// Delete a specific Grade from the Grades table
func (c GradeModel) Delete(id int64) error {

	// check if the id is valid
	if id < 1 {
		return ErrRecordNotFound
	}
	// the SQL query to be executed against the database table
	query := `
        DELETE FROM Grades
        WHERE grade_id = $1`

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
	// gradebably a wrong id was gradevided or the client is trying to
	// delete an already deleted Grade
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil

}

// Get all Grades
func (c GradeModel) GetAll(letter string, filters Filters) ([]*Grade, Metadata, error) {

	query := fmt.Sprintf(`
        SELECT COUNT(*) OVER(), grade_id, created_at, letter_grade, grade_value 
		FROM Grades
		WHERE (to_tsvector('simple', letter_grade) @@
              plainto_tsquery('simple', $1) OR $1 = '') 
		ORDER BY %s %s, grade_id ASC 
		LIMIT $2 OFFSET $3
     `, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// QueryContext returns multiple rows.
	rows, err := c.DB.QueryContext(ctx, query, letter, filters.limit(), filters.offset())
	totalRecords := 0

	// we will store the address of each Grade in our slice
	grades := []*Grade{}

	// gradecess each row that is in rows

	for rows.Next() {
		var grade Grade
		err := rows.Scan(
			&totalRecords,
			&grade.Grade_id,
			&grade.CreatedAt,
			&grade.Grades,
			&grade.Grade_value,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		// add the row to our slice
		grades = append(grades, &grade)
	} // end of for loop
	// after we exit the loop we need to check if it generated any errors
	err = rows.Err()
	if err != nil {
		return nil, Metadata{}, err
	}
	// Create the metadata
	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)

	return grades, metadata, nil

}

// Create a function that performs the validation checks
func ValidateGrades(v *validator.Validator, grade *Grade) {
	// check if the Grade name field is empty
	v.Check(grade.Grades != "", "Letter Grade", "must be provided")

	v.Check(grade.Grade_value <= 4.00, "Grade value", "must be less than or equal to 4.00")
	// check if the Grade code is empty
	v.Check(grade.Grade_value > 0, "Grade Value", "must be greater than zero")

	v.Check(len(grade.Grades) <= 100, "Letter Grade", "must not be more than 100 bytes long")
	// check if the Grade code field is empty
}
