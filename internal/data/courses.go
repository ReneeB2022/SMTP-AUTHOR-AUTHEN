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
type Course struct {
	Course_id     int64     `json:"id"`
	Course_name   string    `json:"Course"` // the Course data
	Course_code   string    `json:"Code"`
	Credits       int64     `json:"credits"`
	Descriptions  string    `json:"description"`
	Prerequisites string    `json:"prereq"`
	Fee           float64   `json:"fee"`
	Dept_id       int64     `json:"-"`
	Dept_code     string    `json:"Category"`
	CreatedAt     time.Time `json:"-"` // database timestamp

}

type CourseModel struct {
	DB *sql.DB
}

// Insert a new row in the departments table
// Expects a pointer to the actual department
func (c CourseModel) Insert(cor *Course) error {
	// the SQL query to be executed against the database table
	query := `
        INSERT INTO Courses (course_name, course_code, credits, Descriptions, Prerequisites, Fee, dept_id)
        VALUES ($1, $2, $3 ,$4, $5, $6, $7)
        RETURNING course_id, created_at
        `
	// the actual values to replace $1, and $2
	args := []any{cor.Course_name, cor.Course_code, cor.Credits, cor.Descriptions,
		cor.Prerequisites, cor.Fee, cor.Dept_id}

	// Create a context with a 3-second timeout. No database
	// operation should take more than 3 seconds or we will quit it
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// execute the query against the departments database table. We ask for the the
	// department_id and created_at to be sent back to us which we will use
	// to update the Department struct later on
	err := c.DB.QueryRowContext(ctx, query, args...).Scan(
		&cor.Course_id,
		&cor.CreatedAt)

	if err != nil {
		return err
	}
	//get remaining empty fields to display
	populatecor, err := c.Get(cor.Course_id)
	if err != nil {
		return err
	}
	cor.Dept_code = populatecor.Dept_code

	return nil

}

// Get a specific Department from the Departments table
func (c CourseModel) Get(id int64) (*Course, error) {
	// check if the id is valid
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	// the SQL query to be executed against the database table
	query := `
        SELECT course_id, Courses.created_at, course_name, course_code, credits, Descriptions, 
		Prerequisites, Fee, dept_id, department_code
        FROM Courses
		INNER JOIN Departments ON Courses.dept_id = Departments.department_id
        WHERE course_id = $1
      `
	// declare a variable of type Departments to store the returned Department
	var cor Course

	// Set a 3-second context/timer
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, id).Scan(
		&cor.Course_id,
		&cor.CreatedAt,
		&cor.Course_name,
		&cor.Course_code,
		&cor.Credits,
		&cor.Descriptions,
		&cor.Prerequisites,
		&cor.Fee,
		&cor.Dept_id,
		&cor.Dept_code,
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
	return &cor, nil
}

// Update a specific Department from the Departments table
func (c CourseModel) Update(cor *Course) error {
	// The SQL query to be executed against the database table
	// Every time we make an update
	query := `
        UPDATE Courses
        SET course_name = $1, course_code = $2, credits = $3, Descriptions = $4, Prerequisites = $5, 
		Fee =$6, dept_id=$7
        WHERE course_id = $8
        RETURNING course_id
      `

	args := []any{cor.Course_name, cor.Course_code, cor.Credits, cor.Descriptions,
		cor.Prerequisites, cor.Fee, cor.Dept_id}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, args...).Scan(&cor.Course_id)

	if err != nil {
		return err
	}
	//get remaining empty fields to display
	populatecor, err := c.Get(cor.Course_id)
	if err != nil {
		return err
	}
	cor.Dept_code = populatecor.Dept_code

	return nil

}

// Delete a specific Department from the Departments table
func (c CourseModel) Delete(id int64) error {

	// check if the id is valid
	if id < 1 {
		return ErrRecordNotFound
	}
	// the SQL query to be executed against the database table
	query := `
        DELETE FROM Courses
        WHERE course_id = $1`

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
func (c CourseModel) GetAll(corname string, corcode string, filters Filters) ([]*Course, Metadata, error) {

	query := fmt.Sprintf(`
        SELECT COUNT(*) OVER(),course_id, Courses.created_at, course_name, course_code, credits, Descriptions, 
		Prerequisites, Fee, dept_id, department_code
        FROM Courses
		INNER JOIN Departments ON Courses.dept_id = Departments.department_id
        WHERE (to_tsvector('simple', course_name) @@
              plainto_tsquery('simple', $1) OR $1 = '') 
        AND (to_tsvector('simple', course_code) @@ 
             plainto_tsquery('simple', $2) OR $2 = '') 
		ORDER BY %s %s, course_id ASC 
		LIMIT $3 OFFSET $4
     `, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// QueryContext returns multiple rows.
	rows, err := c.DB.QueryContext(ctx, query, corname, corcode, filters.limit(), filters.offset())
	totalRecords := 0

	// we will store the address of each Department in our slice
	cors := []*Course{}

	// process each row that is in rows

	for rows.Next() {
		var cor Course
		err := rows.Scan(
			&totalRecords,
			&cor.Course_id,
			&cor.CreatedAt,
			&cor.Course_name,
			&cor.Course_code,
			&cor.Credits,
			&cor.Descriptions,
			&cor.Prerequisites,
			&cor.Fee,
			&cor.Dept_id,
			&cor.Dept_code,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		// add the row to our slice
		cors = append(cors, &cor)
	} // end of for loop
	// after we exit the loop we need to check if it generated any errors
	err = rows.Err()
	if err != nil {
		return nil, Metadata{}, err
	}
	// Create the metadata
	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)

	return cors, metadata, nil

}

// Create a function that performs the validation checks
func ValidateCourse(v *validator.Validator, cor *Course) {
	// check if the Course fields are empty
	v.Check(cor.Course_name != "", "Course name", "must be provided")

	v.Check(cor.Course_code != "", "Course Code", "must be provided")

	v.Check(cor.Prerequisites != "", "prerequisites", "must be provided")

	v.Check(cor.Credits > 0, "Credits", "must be provided")

	v.Check(cor.Descriptions != "", "Description", "must be provided")

	v.Check(cor.Fee > 0, "Fee", "must be provided")

	v.Check(cor.Dept_id > 0, "Department Number", "must be provided")

	v.Check(cor.Dept_id < 5, "Department Id", "must be less than 5")

	// check if the department name field is empty
	v.Check(len(cor.Course_name) <= 100, "Course first name", "must not be more than 100 bytes long")
	// check if the department code field is empty
	v.Check(len(cor.Course_code) <= 25, "Course last name", "must not be more than 25 bytes long")
}
