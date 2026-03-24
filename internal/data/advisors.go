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
type Advisory struct {
	ADSTU_id   int64     `json:"id"`
	Advisor_id int64     `json:"advisor"`
	Student_id int64     `json:"student"` // the Advisory data
	CreatedAt  time.Time `json:"-"`       // database timestamp
}

type AdvisoryModel struct {
	DB *sql.DB
}

// Insert a new row in the Advisorys table
// Expects a pointer to the actual Advisory
func (c AdvisoryModel) Insert(adv *Advisory) error {
	// the SQL query to be executed against the database table
	query := `
        INSERT INTO Advisors_Students (advisor_id, student_id)
        VALUES ($1, $2)
        RETURNING ADSTU_id, created_at
        `
	// the actual values to replace $1, and $2
	args := []any{adv.Advisor_id, adv.Student_id}

	// Create a context with a 3-second timeout. No database
	// operation should take more than 3 seconds or we will quit it
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// execute the query against the Advisorys database table. We ask for the the
	// Advisory_id and created_at to be sent back to us which we will use
	// to update the Advisory struct later on
	return c.DB.QueryRowContext(ctx, query, args...).Scan(
		&adv.ADSTU_id,
		&adv.CreatedAt)
}

// Get a specific Advisory from the Advisorys table
func (c AdvisoryModel) Get(id int64) (*Advisory, error) {
	// check if the id is valid
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	// the SQL query to be executed against the database table
	query := `
        SELECT ADSTU_id, created_at, advisor_id, student_id
        FROM Advisors_Students
        WHERE ADSTU_id = $1
      `
	// declare a variable of type Advisorys to store the returned Advisory
	var adv Advisory

	// Set a 3-second context/timer
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, id).Scan(
		&adv.ADSTU_id,
		&adv.CreatedAt,
		&adv.Advisor_id,
		&adv.Student_id,
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
	return &adv, nil
}

// Update a specific Advisory from the Advisorys table
func (c AdvisoryModel) Update(adv *Advisory) error {
	// The SQL query to be executed against the database table
	// Every time we make an update
	query := `
        UPDATE Advisors_Students
        SET advisor_id = $1
        WHERE ADSTU_id = $2
        RETURNING ADSTU_id
      `

	args := []any{adv.Advisor_id, adv.ADSTU_id}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.DB.QueryRowContext(ctx, query, args...).Scan(&adv.ADSTU_id)

}

// Delete a specific Advisory from the Advisorys table
func (c AdvisoryModel) Delete(id int64) error {

	// check if the id is valid
	if id < 1 {
		return ErrRecordNotFound
	}
	// the SQL query to be executed against the database table
	query := `
        DELETE FROM Advisors_Students
        WHERE ADSTU_id = $1`

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
	// delete an already deleted Advisory
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil

}

// Get all Advisorys
func (c AdvisoryModel) GetAll(advid int64, filters Filters) ([]*Advisory, Metadata, error) {

	query := fmt.Sprintf(`
        SELECT COUNT(*) OVER(),ADSTU_id, created_at, student_id, advisor_id
        FROM Advisors_Students
        WHERE (advisor_id = $1 OR $1 = 0)
		ORDER BY %s %s, ADSTU_id ASC 
		LIMIT $2 OFFSET $3
     `, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// QueryContext returns multiple rows.
	rows, err := c.DB.QueryContext(ctx, query, advid, filters.limit(), filters.offset())
	totalRecords := 0

	// we will store the address of each Advisory in our slice
	advises := []*Advisory{}

	// process each row that is in rows

	for rows.Next() {
		var advise Advisory
		err := rows.Scan(
			&totalRecords,
			&advise.ADSTU_id,
			&advise.CreatedAt,
			&advise.Student_id,
			&advise.Advisor_id,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		// add the row to our slice
		advises = append(advises, &advise)
	} // end of for loop
	// after we exit the loop we need to check if it generated any errors
	err = rows.Err()
	if err != nil {
		return nil, Metadata{}, err
	}
	// Create the metadata
	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)

	return advises, metadata, nil

}

// Create a function that performs the validation checks
func ValidateAdvisors(v *validator.Validator, advise *Advisory) {
	// check if the Advisory name field is empty
	v.Check(advise.Student_id > 0, "Student", "must be provided")
	// check if the Advisory code is empty
	v.Check(advise.Advisor_id > 0, "Advisor", "must be provided")
}
