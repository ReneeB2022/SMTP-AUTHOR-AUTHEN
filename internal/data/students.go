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
type Student struct {
	Student_id    int64     `json:"id"`
	Fname         string    `json:"Fname"` // the student data
	Lname         string    `json:"Lname"`
	Gender        string    `json:"gender"`
	Age           int64     `json:"age"`
	District      int64     `json:"-"`
	District_name string    `json:"district_name"`
	Program_id    int64     `json:"-"`
	Program_code  string    `json:"program code"`
	GPA           float64   `json:"gpa"`
	CreatedAt     time.Time `json:"-"` // database timestamp
}

type StudentModel struct {
	DB *sql.DB
}

// Insert a new row in the departments table
// Expects a pointer to the actual department
func (c StudentModel) Insert(stu *Student) error {
	// the SQL query to be executed against the database table
	query := `
        INSERT INTO Students (Fname, Lname, Gender, age, District, program_id, GPA)
        VALUES ($1, $2, $3 ,$4, $5, $6, $7)
        RETURNING student_id, created_at
        `
	// the actual values to replace $1, and $2
	args := []any{stu.Fname, stu.Lname, stu.Gender, stu.Age, stu.District, stu.Program_id, stu.GPA}

	// Create a context with a 3-second timeout. No database
	// operation should take more than 3 seconds or we will quit it
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// execute the query against the departments database table. We ask for the the
	// department_id and created_at to be sent back to us which we will use
	// to update the Department struct later on
	err := c.DB.QueryRowContext(ctx, query, args...).Scan(
		&stu.Student_id,
		&stu.CreatedAt)

	if err != nil {
		return err
	}

	populatestu, err := c.Get(stu.Student_id)
	if err != nil {
		return err
	}
	stu.District_name = populatestu.District_name
	stu.Program_code = populatestu.Program_code

	return nil
}

// Get a specific Department from the Departments table
func (c StudentModel) Get(id int64) (*Student, error) {
	// check if the id is valid
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	// the SQL query to be executed against the database table
	query := `
        SELECT student_id, Students.created_at, Fname, Lname, Gender, Age, District, district_name, 
		Students.program_id, program_code, GPA
        FROM Students
		INNER JOIN Districts ON Students.District = Districts.district_id
		INNER JOIN Programs ON Students.program_id = Programs.program_id
        WHERE student_id = $1
      `
	// declare a variable of type Departments to store the returned Department
	var stu Student

	// Set a 3-second context/timer
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, id).Scan(
		&stu.Student_id,
		&stu.CreatedAt,
		&stu.Fname,
		&stu.Lname,
		&stu.Gender,
		&stu.Age,
		&stu.District,
		&stu.District_name,
		&stu.Program_id,
		&stu.Program_code,
		&stu.GPA,
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
	return &stu, nil
}

// Update a specific Department from the Departments table
func (c StudentModel) Update(stu *Student) error {
	// The SQL query to be executed against the database table
	// Every time we make an update
	query := `
        UPDATE Students
        SET Fname = $1, Lname = $2, Gender = $3, age = $4, District = $5, program_id = $6, GPA =$7
        WHERE student_id = $8
        RETURNING student_id
      `

	args := []any{stu.Fname, stu.Lname, stu.Gender, stu.Age, stu.District, stu.Program_id, stu.GPA, stu.Student_id}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, args...).Scan(&stu.Student_id)

	if err != nil {
		return err
	}

	populatestu, err := c.Get(stu.Student_id)
	if err != nil {
		return err
	}
	stu.District_name = populatestu.District_name
	stu.Program_code = populatestu.Program_code

	return nil

}

// Delete a specific Department from the Departments table
func (c StudentModel) Delete(id int64) error {

	// check if the id is valid
	if id < 1 {
		return ErrRecordNotFound
	}
	// the SQL query to be executed against the database table
	query := `
        DELETE FROM Students
        WHERE student_id = $1`

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
func (c StudentModel) GetAll(Fname string, Lname string, District int64, filters Filters) ([]*Student, Metadata, error) {

	query := fmt.Sprintf(`
        SELECT COUNT(*) OVER(),student_id, Students.created_at, Fname, Lname, District, district_name, Gender,
		age, Students.program_id, program_code, GPA
        FROM Students
		INNER JOIN Districts ON Students.District = Districts.district_id
		INNER JOIN Programs ON Students.program_id = Programs.program_id
        WHERE (to_tsvector('simple', Fname) @@
              plainto_tsquery('simple', $1) OR $1 = '') 
        AND (to_tsvector('simple', Lname) @@ 
             plainto_tsquery('simple', $2) OR $2 = '') 
		AND (District = $3 OR $3 = 0)
		ORDER BY %s %s, student_id ASC 
		LIMIT $4 OFFSET $5
     `, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// QueryContext returns multiple rows.
	rows, err := c.DB.QueryContext(ctx, query, Fname, Lname, District, filters.limit(), filters.offset())
	totalRecords := 0

	// we will store the address of each Department in our slice
	stus := []*Student{}

	// process each row that is in rows

	for rows.Next() {
		var stu Student
		err := rows.Scan(
			&totalRecords,
			&stu.Student_id,
			&stu.CreatedAt,
			&stu.Fname,
			&stu.Lname,
			&stu.District,
			&stu.District_name,
			&stu.Gender,
			&stu.Age,
			&stu.Program_id,
			&stu.Program_code,
			&stu.GPA,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		// add the row to our slice
		stus = append(stus, &stu)
	} // end of for loop
	// after we exit the loop we need to check if it generated any errors
	err = rows.Err()
	if err != nil {
		return nil, Metadata{}, err
	}
	// Create the metadata
	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)

	return stus, metadata, nil

}

// Create a function that performs the validation checks
func Validatestudent(v *validator.Validator, stu *Student) {
	// check if the student fields are empty
	v.Check(stu.Fname != "", "First name", "must be provided")

	v.Check(stu.Lname != "", "Last name", "must be provided")

	v.Check(stu.Gender != "", "Gender", "must be provided")

	v.Check(stu.Age > 4, "Valid age", "must be provided")

	v.Check(stu.District > 0, "District code", "must be provided")

	v.Check(stu.GPA >= 0, "Valid GPA", "must be provided, cannot be less than zero")

	v.Check(stu.GPA <= 4.00, "GPA", "cannot exceed 4.00")

	v.Check(stu.Program_id > 0, "Program code", "must be provided")
	// check if the department name field is empty
	v.Check(len(stu.Fname) <= 100, "Student first name", "must not be more than 100 bytes long")
	// check if the department code field is empty
	v.Check(len(stu.Lname) <= 25, "Student last name", "must not be more than 25 bytes long")
}
