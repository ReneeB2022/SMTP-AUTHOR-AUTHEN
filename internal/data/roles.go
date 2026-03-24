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
type Role struct {
	Roles_id  int64     `json:"id"`
	Roles     string    `json:"role"` // the Role data
	CreatedAt time.Time `json:"-"`    // database timestamp
}

type RoleModel struct {
	DB *sql.DB
}

// Insert a new row in the Roles table
// Expects a pointer to the actual Role
func (c RoleModel) Insert(role *Role) error {
	// the SQL query to be executed against the database table
	query := `
        INSERT INTO Roles (roles)
        VALUES ($1)
        RETURNING role_id, created_at
        `
	// the actual values to replace $1, and $2
	args := []any{role.Roles}

	// Create a context with a 3-second timeout. No database
	// operation should take more than 3 seconds or we will quit it
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// execute the query against the Roles database table. We ask for the the
	// Role_id and created_at to be sent back to us which we will use
	// to update the Role struct later on
	return c.DB.QueryRowContext(ctx, query, args...).Scan(
		&role.Roles_id,
		&role.CreatedAt)
}

// Get a specific Role from the Roles table
func (c RoleModel) Get(id int64) (*Role, error) {
	// check if the id is valid
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	// the SQL query to be executed against the database table
	query := `
        SELECT role_id, created_at, roles
        FROM Roles
        WHERE role_id = $1
      `
	// declare a variable of type Roles to store the returned Role
	var ro Role

	// Set a 3-second context/timer
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, id).Scan(
		&ro.Roles_id,
		&ro.CreatedAt,
		&ro.Roles,
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
	return &ro, nil
}

// Update a specific Role from the Roles table
func (c RoleModel) Update(ro *Role) error {
	// The SQL query to be executed against the database table
	// Every time we make an update
	query := `
        UPDATE Roles
        SET roles = $1
        WHERE role_id = $2
        RETURNING role_id
      `

	args := []any{ro.Roles, ro.Roles_id}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.DB.QueryRowContext(ctx, query, args...).Scan(&ro.Roles_id)

}

// Delete a specific Role from the Roles table
func (c RoleModel) Delete(id int64) error {

	// check if the id is valid
	if id < 1 {
		return ErrRecordNotFound
	}
	// the SQL query to be executed against the database table
	query := `
        DELETE FROM Roles
        WHERE role_id = $1`

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
	// delete an already deleted Role
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil

}

// Get all Roles
func (c RoleModel) GetAll(ro string, filters Filters) ([]*Role, Metadata, error) {

	query := fmt.Sprintf(`
        SELECT COUNT(*) OVER(),role_id, created_at, roles
        FROM Roles
        WHERE (to_tsvector('simple', roles) @@
              plainto_tsquery('simple', $1) OR $1 = '') 
		ORDER BY %s %s, role_id ASC 
		LIMIT $2 OFFSET $3
     `, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// QueryContext returns multiple rows.
	rows, err := c.DB.QueryContext(ctx, query, ro, filters.limit(), filters.offset())
	totalRecords := 0

	// we will store the address of each Role in our slice
	roles := []*Role{}

	// process each row that is in rows

	for rows.Next() {
		var rol Role
		err := rows.Scan(
			&totalRecords,
			&rol.Roles_id,
			&rol.CreatedAt,
			&rol.Roles,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		// add the row to our slice
		roles = append(roles, &rol)
	} // end of for loop
	// after we exit the loop we need to check if it generated any errors
	err = rows.Err()
	if err != nil {
		return nil, Metadata{}, err
	}
	// Create the metadata
	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)

	return roles, metadata, nil

}

// Create a function that performs the validation checks
func ValidateRole(v *validator.Validator, ro *Role) {
	// check if the Role name field is empty
	v.Check(ro.Roles != "", "Role code", "must be provided")
	// check if the Role name field is empty
	v.Check(len(ro.Roles) <= 100, "Role name", "must not be more than 100 bytes long")
	// check if the Role code field is empty
	//v.Check(len(dept.Role_code) <= 25, "Role code", "must not be more than 25 bytes long")
}
