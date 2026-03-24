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
type District struct {
	District_id   int64     `json:"id"`
	District_name string    `json:"district"` // the District data
	CreatedAt     time.Time `json:"-"`        // database timestamp
}

type DistrictModel struct {
	DB *sql.DB
}

// Insert a new row in the Districts table
// Expects a pointer to the actual District
func (c DistrictModel) Insert(district *District) error {
	// the SQL query to be executed against the database table
	query := `
        INSERT INTO Districts (district_name)
        VALUES ($1)
        RETURNING district_id, created_at
        `
	// the actual values to replace $1, and $2
	args := []any{district.District_name}

	// Create a context with a 3-second timeout. No database
	// operation should take more than 3 seconds or we will quit it
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// execute the query against the Districts database table. We ask for the the
	// District_id and created_at to be sent back to us which we will use
	// to update the District struct later on
	return c.DB.QueryRowContext(ctx, query, args...).Scan(
		&district.District_id,
		&district.CreatedAt)
}

// Get a specific District from the Districts table
func (c DistrictModel) Get(id int64) (*District, error) {
	// check if the id is valid
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	// the SQL query to be executed against the database table
	query := `
        SELECT district_id, created_at, district_name
        FROM Districts
        WHERE district_id = $1
      `
	// declare a variable of type Districts to store the returned District
	var dis District

	// Set a 3-second context/timer
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, id).Scan(
		&dis.District_id,
		&dis.CreatedAt,
		&dis.District_name,
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
	return &dis, nil
}

// Update a specific District from the Districts table
func (c DistrictModel) Update(dis *District) error {
	// The SQL query to be executed against the database table
	// Every time we make an update
	query := `
        UPDATE Districts
        SET district_name = $1
        WHERE district_id = $2
        RETURNING district_id
      `

	args := []any{dis.District_name, dis.District_id}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.DB.QueryRowContext(ctx, query, args...).Scan(&dis.District_id)

}

// Delete a specific District from the Districts table
func (c DistrictModel) Delete(id int64) error {

	// check if the id is valid
	if id < 1 {
		return ErrRecordNotFound
	}
	// the SQL query to be executed against the database table
	query := `
        DELETE FROM Districts
        WHERE district_id = $1`

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
	// delete an already deleted District
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil

}

// Get all Districts
func (c DistrictModel) GetAll(dis string, filters Filters) ([]*District, Metadata, error) {

	query := fmt.Sprintf(`
        SELECT COUNT(*) OVER(),district_id, created_at, district_name
        FROM Districts
        WHERE (to_tsvector('simple', district_name) @@
              plainto_tsquery('simple', $1) OR $1 = '') 
		ORDER BY %s %s, district_id ASC 
		LIMIT $2 OFFSET $3
     `, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// QueryContext returns multiple rows.
	rows, err := c.DB.QueryContext(ctx, query, dis, filters.limit(), filters.offset())
	totalRecords := 0

	// we will store the address of each District in our slice
	districts := []*District{}

	// process each row that is in rows

	for rows.Next() {
		var district District
		err := rows.Scan(
			&totalRecords,
			&district.District_id,
			&district.CreatedAt,
			&district.District_name,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		// add the row to our slice
		districts = append(districts, &district)
	} // end of for loop
	// after we exit the loop we need to check if it generated any errors
	err = rows.Err()
	if err != nil {
		return nil, Metadata{}, err
	}
	// Create the metadata
	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)

	return districts, metadata, nil

}

// Create a function that performs the validation checks
func ValidateDistrict(v *validator.Validator, dis *District) {
	// check if the District name field is empty
	v.Check(dis.District_name != "", "District name", "must be provided")
	// check if the District name field is empty
	v.Check(len(dis.District_name) <= 100, "District name", "must not be more than 100 bytes long")
	// check if the District code field is empty
	//v.Check(len(dept.District_code) <= 25, "District code", "must not be more than 25 bytes long")
}
