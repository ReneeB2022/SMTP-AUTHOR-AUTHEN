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
type Staff struct {
	Staff_id        int64     `json:"id"`
	Fname           string    `json:"Fname"` // the student data
	Lname           string    `json:"Lname"`
	Role_id         int64     `json:"-"`
	Role_name       string    `json:"role"`
	Depart_Id       int64     `json:"-"`
	Department_name string    `json:"Faculty"`
	CreatedAt       time.Time `json:"-"` // database timestamp
}

type StaffModel struct {
	DB *sql.DB
}

// Insert a new row in the departments table
// Expects a pointer to the actual department
func (c StaffModel) Insert(staff *Staff) error {
	// the SQL query to be executed against the database table
	query := `
        INSERT INTO Staff (Fname, Lname, Role_id, Depart_id)
        VALUES ($1, $2, $3 ,$4)
        RETURNING staff_id, created_at
        `
	// the actual values to replace $1, and $2
	args := []any{staff.Fname, staff.Lname, staff.Role_id, staff.Depart_Id}

	// Create a context with a 3-second timeout. No database
	// operation should take more than 3 seconds or we will quit it
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// execute the query against the departments database table. We ask for the the
	// department_id and created_at to be sent back to us which we will use
	// to update the Department struct later on
	err := c.DB.QueryRowContext(ctx, query, args...).Scan(
		&staff.Staff_id,
		&staff.CreatedAt)

	if err != nil {
		return err
	}
	//get remaining empty fields to display
	populatestaff, err := c.Get(staff.Staff_id)
	if err != nil {
		return err
	}
	staff.Role_name = populatestaff.Role_name
	staff.Department_name = populatestaff.Department_name

	return nil
}

// Get a specific Department from the Departments table
func (c StaffModel) Get(id int64) (*Staff, error) {
	// check if the id is valid
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	// the SQL query to be executed against the database table
	query := `
        SELECT staff_id, Staff.created_at, Fname, Lname, Staff.Role_id, Depart_id ,roles, department_code
        FROM Staff
		INNER JOIN Roles ON Staff.Role_id = Roles.role_id
		INNER JOIN Departments ON Staff.Depart_id = Departments.department_id
        WHERE staff_id = $1
      `
	// declare a variable of type Departments to store the returned Department
	var staff Staff

	// Set a 3-second context/timer
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, id).Scan(
		&staff.Staff_id,
		&staff.CreatedAt,
		&staff.Fname,
		&staff.Lname,
		&staff.Role_id,
		&staff.Depart_Id,
		&staff.Role_name,
		&staff.Department_name,
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
	return &staff, nil
}

// Update a specific Department from the Departments table
func (c StaffModel) Update(staff *Staff) error {
	// The SQL query to be executed against the database table
	// Every time we make an update
	query := `
        UPDATE Staff
        SET Fname = $1, Lname = $2, Role_id = $3, Depart_id = $4
        WHERE staff_id = $5
        RETURNING staff_id
      `

	args := []any{staff.Fname, staff.Lname, staff.Role_id, staff.Depart_Id, staff.Staff_id}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, args...).Scan(&staff.Staff_id)
	if err != nil {
		return err
	}
	//get remaining empty fields to display
	updatestaff, err := c.Get(staff.Staff_id)
	if err != nil {
		return err
	}
	staff.Department_name = updatestaff.Department_name
	staff.Role_name = updatestaff.Role_name

	return nil

}

// Delete a specific Department from the Departments table
func (c StaffModel) Delete(id int64) error {

	// check if the id is valid
	if id < 1 {
		return ErrRecordNotFound
	}
	// the SQL query to be executed against the database table
	query := `
        DELETE FROM Staff
        WHERE staff_id = $1`

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
func (c StaffModel) GetAll(Fname string, Lname string, filters Filters) ([]*Staff, Metadata, error) {

	query := fmt.Sprintf(`
        SELECT COUNT(*) OVER(),staff_id, Staff.created_at, Fname, Lname, Staff.Role_id, Depart_id,
		roles, department_code
        FROM Staff
		INNER JOIN Roles ON Staff.Role_id = Roles.role_id
		INNER JOIN Departments ON Staff.Depart_id = Departments.department_id
	    WHERE (to_tsvector('simple', Fname) @@
              plainto_tsquery('simple', $1) OR $1 = '') 
        AND (to_tsvector('simple', Lname) @@ 
             plainto_tsquery('simple', $2) OR $2 = '') 
		ORDER BY %s %s, staff_id ASC 
		LIMIT $3 OFFSET $4
     `, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// QueryContext returns multiple rows.
	rows, err := c.DB.QueryContext(ctx, query, Fname, Lname, filters.limit(), filters.offset())
	totalRecords := 0

	// we will store the address of each Department in our slice
	staffs := []*Staff{}

	// process each row that is in rows

	for rows.Next() {
		var staff Staff
		err := rows.Scan(
			&totalRecords,
			&staff.Staff_id,
			&staff.CreatedAt,
			&staff.Fname,
			&staff.Lname,
			&staff.Role_id,
			&staff.Depart_Id,
			&staff.Role_name,
			&staff.Department_name,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		// add the row to our slice
		staffs = append(staffs, &staff)
	} // end of for loop
	// after we exit the loop we need to check if it generated any errors
	err = rows.Err()
	if err != nil {
		return nil, Metadata{}, err
	}
	// Create the metadata
	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)

	return staffs, metadata, nil

}

// Create a function that performs the validation checks
func Validatestaff(v *validator.Validator, staff *Staff) {
	// check if the student fields are empty
	v.Check(staff.Fname != "", "First name", "must be provided")

	v.Check(staff.Lname != "", "Last name", "must be provided")

	v.Check(staff.Role_id > 0, "Role id", "must be provided")

	v.Check(staff.Depart_Id > 0, "Department id", "must be provided")

	// check if the department name field is empty
	v.Check(len(staff.Fname) <= 100, "Staff member first name", "must not be more than 100 bytes long")
	// check if the department code field is empty
	v.Check(len(staff.Lname) <= 25, "Staff member last name", "must not be more than 25 bytes long")
}
