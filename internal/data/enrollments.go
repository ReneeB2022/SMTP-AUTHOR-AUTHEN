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
type Enrollment struct {
	Enrollment_id int64     `json:"id"`
	Student_id    int64     `json:"student"` // the Enrollment data
	Section_id    int64     `json:"section"`
	Grade_id      int64     `json:"grade"`
	Date          time.Time `json:"enrolled"`
	CreatedAt     time.Time `json:"-"` // database timestamp
}

type EnrollmentModel struct {
	DB *sql.DB
}

// Insert a new row in the Enrollments table
// Expects a pointer to the actual Enrollment
func (c EnrollmentModel) Insert(enroll *Enrollment) error {
	// the SQL query to be executed against the database table
	query := `
        INSERT INTO Enrollments (student_id, section_id, enrollment_date, grade_id)
        VALUES ($1, $2, $3, $4)
        RETURNING enrollment_id, created_at
        `
	// the actual values to replace $1, and $2
	args := []any{enroll.Student_id, enroll.Section_id, enroll.Date, enroll.Grade_id}

	// Create a context with a 3-second timeout. No database
	// operation should take more than 3 seconds or we will quit it
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// execute the query against the Enrollments database table. We ask for the the
	// Enrollment_id and created_at to be sent back to us which we will use
	// to update the Enrollment struct later on
	return c.DB.QueryRowContext(ctx, query, args...).Scan(
		&enroll.Enrollment_id,
		&enroll.CreatedAt)
}

// Get a specific Enrollment from the Enrollments table
func (c EnrollmentModel) Get(id int64) (*Enrollment, error) {
	// check if the id is valid
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	// the SQL query to be executed against the database table
	query := `
        SELECT enrollment_id, created_at, student_id, section_id, enrollment_date, grade_id
        FROM Enrollments
        WHERE enrollment_id = $1
      `
	// declare a variable of type Enrollments to store the returned Enrollment
	var enroll Enrollment

	// Set a 3-second context/timer
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, id).Scan(
		&enroll.Enrollment_id,
		&enroll.CreatedAt,
		&enroll.Student_id,
		&enroll.Section_id,
		&enroll.Date,
		&enroll.Grade_id,
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
	return &enroll, nil
}

// Update a specific Enrollment from the Enrollments table
func (c EnrollmentModel) Update(enroll *Enrollment) error {
	// The SQL query to be executed against the database table
	// Every time we make an update
	query := `
        UPDATE Enrollments
        SET grade_id = $1
        WHERE enrollment_id = $2
        RETURNING enrollment_id
      `

	args := []any{enroll.Grade_id, enroll.Enrollment_id}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.DB.QueryRowContext(ctx, query, args...).Scan(&enroll.Enrollment_id)

}

// Delete a specific Enrollment from the Enrollments table
func (c EnrollmentModel) Delete(id int64) error {

	// check if the id is valid
	if id < 1 {
		return ErrRecordNotFound
	}
	// the SQL query to be executed against the database table
	query := `
        DELETE FROM Enrollments
        WHERE enrollment_id = $1`

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
	// delete an already deleted Enrollment
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil

}

// Get all Enrollments
func (c EnrollmentModel) GetAll(sectionid int64, grade_id int64, filters Filters) ([]*Enrollment, Metadata, error) {

	query := fmt.Sprintf(`
        SELECT COUNT(*) OVER(),enrollment_id, created_at, student_id, section_id, enrollment_date, grade_id
        FROM Enrollments
        WHERE (section_id = $1 OR $1 = 0)
        AND (grade_id = $2 OR $2 = 0)
		ORDER BY %s %s, Enrollment_id ASC 
		LIMIT $3 OFFSET $4
     `, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// QueryContext returns multiple rows.
	rows, err := c.DB.QueryContext(ctx, query, sectionid, grade_id, filters.limit(), filters.offset())
	totalRecords := 0

	// we will store the address of each Enrollment in our slice
	enrolls := []*Enrollment{}

	// process each row that is in rows

	for rows.Next() {
		var enroll Enrollment
		err := rows.Scan(
			&totalRecords,
			&enroll.Enrollment_id,
			&enroll.CreatedAt,
			&enroll.Student_id,
			&enroll.Section_id,
			&enroll.Date,
			&enroll.Grade_id,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		// add the row to our slice
		enrolls = append(enrolls, &enroll)
	} // end of for loop
	// after we exit the loop we need to check if it generated any errors
	err = rows.Err()
	if err != nil {
		return nil, Metadata{}, err
	}
	// Create the metadata
	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)

	return enrolls, metadata, nil

}

// Create a function that performs the validation checks
func ValidateEnrollment(v *validator.Validator, enroll *Enrollment) {
	// check if the Enrollment code is empty
	v.Check(enroll.Student_id > 0, "Enrollment code", "must be provided")

	v.Check(enroll.Grade_id >= 0, "Grade_id", "must be provided")

	v.Check(enroll.Section_id > 0, "Section Id", "must be Valid")

	v.Check(!enroll.Date.IsZero(), "Date", "Must be Provided")

}
