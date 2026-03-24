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
type Section struct {
	Section_id   int64     `json:"id"`
	Staff_id     int64     `json:"-"`
	First        string    `json:"Lecturer_First"`
	Last         string    `json:"Lecturer_Last"`
	Course_id    int64     `json:"-"`
	Course       string    `json:"Course"`
	Availability int64     `json:"available"`
	Classroom    string    `json:"Room"`
	Classday     string    `json:"Day"`
	Start        time.Time `json:"starts"`
	End          time.Time `json:"finish"`
	Semester     string    `json:"semester"`
	CreatedAt    time.Time `json:"-"` // database timestamp
}

type SectionModel struct {
	DB *sql.DB
}

// Insert a new row in the departments table
// Expects a pointer to the actual department
func (c SectionModel) Insert(sec *Section) error {
	// the SQL query to be executed against the database table
	query := `
        INSERT INTO Sections (staff_id, course_id, availability, classroom, classday, starttime, endtime, semester)
        VALUES ($1, $2, $3 ,$4, $5, $6, $7, $8)
        RETURNING section_id, created_at
        `
	// the actual values to replace $1, and $2
	args := []any{sec.Staff_id, sec.Course_id, sec.Availability, sec.Classroom, sec.Classday, sec.Start,
		sec.End, sec.Semester}

	// Create a context with a 3-second timeout. No database
	// operation should take more than 3 seconds or we will quit it
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// execute the query against the departments database table. We ask for the the
	// department_id and created_at to be sent back to us which we will use
	// to update the Department struct later on
	err := c.DB.QueryRowContext(ctx, query, args...).Scan(
		&sec.Section_id,
		&sec.CreatedAt)

	if err != nil {
		return err
	}

	populatesec, err := c.Get(sec.Section_id)
	if err != nil {
		return err
	}
	sec.First = populatesec.First
	sec.Last = populatesec.Last
	sec.Course = populatesec.Course

	return nil
}

// Get a specific Department from the Departments table
func (c SectionModel) Get(id int64) (*Section, error) {
	// check if the id is valid
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	// the SQL query to be executed against the database table
	query := `
        SELECT section_id, Sections.created_at, Sections.staff_id, Sections.course_id, availability, classroom, classday,
		starttime, endtime, semester, Fname, Lname, course_name
        FROM Sections
		INNER JOIN Staff ON Sections.staff_id = Staff.staff_id
		INNER JOIN Courses ON Sections.course_id = Courses.course_id
        WHERE section_id = $1
      `
	// declare a variable of type Departments to store the returned Department
	var sec Section

	// Set a 3-second context/timer
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, id).Scan(
		&sec.Section_id,
		&sec.CreatedAt,
		&sec.Staff_id,
		&sec.Course_id,
		&sec.Availability,
		&sec.Classroom,
		&sec.Classday,
		&sec.Start,
		&sec.End,
		&sec.Semester,
		&sec.First,
		&sec.Last,
		&sec.Course,
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
	return &sec, nil
}

// Update a specific Department from the Departments table
func (c SectionModel) Update(sec *Section) error {
	// The SQL query to be executed against the database table
	// Every time we make an update
	query := `
        UPDATE Sections
        SET staff_id= $1, course_id = $2, availability = $3, classroom = $4, classday = $5, starttime = $6, 
		endtime =$7, semester=$8
        WHERE section_id = $9
        RETURNING section_id
      `

	args := []any{sec.Staff_id, sec.Course_id, sec.Availability, sec.Classroom, sec.Classday, sec.Start,
		sec.End, sec.Semester}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, args...).Scan(&sec.Section_id)

	if err != nil {
		return err
	}

	populatesec, err := c.Get(sec.Section_id)
	if err != nil {
		return err
	}
	sec.First = populatesec.First
	sec.Last = populatesec.Last
	sec.Course = populatesec.Course

	return nil

}

// Delete a specific Department from the Departments table
func (c SectionModel) Delete(id int64) error {

	// check if the id is valid
	if id < 1 {
		return ErrRecordNotFound
	}
	// the SQL query to be executed against the database table
	query := `
        DELETE FROM Sections
        WHERE section_id = $1`

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
func (c SectionModel) GetAll(courseid int64, filters Filters) ([]*Section, Metadata, error) {

	query := fmt.Sprintf(`
        SELECT COUNT(*) OVER(), section_id, Sections.created_at, Sections.staff_id, Sections.course_id, availability, 
		classroom, classday, starttime, endtime, semester, Fname, Lname, course_name
        FROM Sections
		INNER JOIN Staff ON Sections.staff_id = Staff.staff_id
		INNER JOIN Courses ON Sections.course_id = Courses.course_id
        WHERE (Sections.course_id= $1 OR $1 = 0)
		ORDER BY %s %s, section_id ASC 
		LIMIT $2 OFFSET $3
     `, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// QueryContext returns multiple rows.
	rows, err := c.DB.QueryContext(ctx, query, courseid, filters.limit(), filters.offset())
	totalRecords := 0

	// we will store the address of each Department in our slice
	secs := []*Section{}

	// process each row that is in rows

	for rows.Next() {
		var sec Section
		err := rows.Scan(
			&totalRecords,
			&sec.Section_id,
			&sec.CreatedAt,
			&sec.Staff_id,
			&sec.Course_id,
			&sec.Availability,
			&sec.Classroom,
			&sec.Classday,
			&sec.Start,
			&sec.End,
			&sec.Semester,
			&sec.First,
			&sec.Last,
			&sec.Course,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		// add the row to our slice
		secs = append(secs, &sec)
	} // end of for loop
	// after we exit the loop we need to check if it generated any errors
	err = rows.Err()
	if err != nil {
		return nil, Metadata{}, err
	}
	// Create the metadata
	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)

	return secs, metadata, nil

}

// Create a function that performs the validation checks
func ValidateSection(v *validator.Validator, sec *Section) {
	// check if the Section fields are empty
	v.Check(sec.Staff_id > 0, "Staff member ID", "must be provided")

	v.Check(sec.Course_id > 0, "course ID", "must be provided")

	v.Check(sec.Availability > 0, "Available Seats", "must be provided")

	v.Check(sec.Classroom != "", "Classroom", "must be provided")

	v.Check(sec.Classday != "", "Class Day", "must be provided")

	v.Check(sec.Semester != "", "Semester", "must be provided")

	v.Check(sec.Start.Before(sec.End), "start time", "cannot be occur after end time")

	v.Check(sec.End.After(sec.Start), "End time", "cannot occur before start time")

	// check if the department name field is empty
	v.Check(len(sec.Classroom) <= 100, "Semester", "must not be more than 100 bytes long")

	v.Check(len(sec.Classday) <= 100, "Semester", "must not be more than 100 bytes long")
	// check if the department code field is empty
	v.Check(len(sec.Semester) <= 25, "Section last name", "must not be more than 25 bytes long")
}
