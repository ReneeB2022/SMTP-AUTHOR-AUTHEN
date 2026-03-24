package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ReneeB2022/test1/internal/data"
	"github.com/ReneeB2022/test1/internal/validator"
)

func (a *applicationDependencies) createstaffHandler(w http.ResponseWriter,
	r *http.Request) {
	// create a struct to hold a Department
	// we use struct tags[``] to make the names display in lowercase
	var incomingData struct {
		Fname     string `json:"Fname"` // the student data
		Lname     string `json:"Lname"`
		Role_id   int64  `json:"role"`
		Depart_Id int64  `json:"department"`
	}
	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.errorResponseJSON(w, r, http.StatusBadRequest, err.Error())
		return
	}
	staff := &data.Staff{
		Fname:     incomingData.Fname,
		Lname:     incomingData.Lname,
		Role_id:   incomingData.Role_id,
		Depart_Id: incomingData.Depart_Id,
	}
	// Initialize a Validator instance
	v := validator.New()

	data.Validatestaff(v, staff)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors) // implemented later
		return
	}

	err = a.staffModel.Insert(staff)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/staff/%d", staff.Staff_id))

	// Send a JSON response with 201 (new resource created) status code
	data := envelope{
		"Staff": staff,
	}
	err = a.writeJSON(w, http.StatusCreated, data, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

}
func (a *applicationDependencies) displaystaffHandler(
	w http.ResponseWriter,
	r *http.Request) {

	// Get the id from the URL /v1/departments/:id so that we
	// can use it to query teh Departments table. We will
	// implement the readIDParam() function later
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}
	// Call Get() to retrieve the Department with the specified id
	staff, err := a.staffModel.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}
	// display the Department
	data := envelope{
		"Staff": staff,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

}
func (a *applicationDependencies) updatestaffHandler(
	w http.ResponseWriter,
	r *http.Request) {
	// Get the id from the URL
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	// Call Get() to retrieve the Department with the specified id
	staff, err := a.staffModel.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	// Use our temporary incomingData struct to hold the data
	// Note: I have changed the types to pointer to differentiate
	// between the client leaving a field empty intentionally
	// and the field not needing to be updated
	var incomingData struct {
		Fname     string `json:"Fname"` // the student data
		Lname     string `json:"Lname"`
		Role_id   *int64 `json:"role"`
		Depart_Id *int64 `json:"department"`
	}

	// perform the decoding
	err = a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}
	// We need to now check the fields to see which ones need updating
	// if incomingData.Department_name is nil, no update was provided
	if incomingData.Fname != "" {
		staff.Fname = incomingData.Fname
	}
	// if incomingData.Department_code is nil, no update was provided
	if incomingData.Lname != "" {
		staff.Lname = incomingData.Lname
	}
	if incomingData.Role_id != nil {
		staff.Role_id = *incomingData.Role_id
	}
	if incomingData.Depart_Id != nil {
		staff.Depart_Id = *incomingData.Depart_Id
	}

	// Before we write the updates to the DB let's validate
	v := validator.New()
	data.Validatestaff(v, staff)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}
	// perform the update
	err = a.staffModel.Update(staff)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	data := envelope{
		"Staff": staff,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}
func (a *applicationDependencies) deletestaffHandler(
	w http.ResponseWriter,
	r *http.Request) {

	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	err = a.staffModel.Delete(id)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	// display the Department
	data := envelope{
		"message": "Staff member successfully deleted",
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}

}

func (a *applicationDependencies) liststaffHandler(
	w http.ResponseWriter,
	r *http.Request) {

	var queryParametersData struct {
		Fname string
		Lname string
		data.Filters
	}

	queryParameters := r.URL.Query()

	queryParametersData.Fname = a.getSingleQueryParameter(
		queryParameters,
		"Fname",
		"")

	queryParametersData.Lname = a.getSingleQueryParameter(
		queryParameters,
		"Lname",
		"")

	v := validator.New()

	queryParametersData.Filters.Page = a.getSingleIntegerParameter(
		queryParameters, "page", 1, v)
	queryParametersData.Filters.PageSize = a.getSingleIntegerParameter(
		queryParameters, "page_size", 10, v)

	queryParametersData.Filters.Sort = a.getSingleQueryParameter(
		queryParameters, "sort", "staff_id")

	queryParametersData.Filters.SortSafeList = []string{"staff_id", "Fname", "Lname",
		"-staff_id", "-Fname", "-Lname"}

	staff, metadata, err := a.staffModel.GetAll(queryParametersData.Fname,
		queryParametersData.Lname, queryParametersData.Filters)

	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{
		"Staff":     staff,
		"@metadata": metadata,
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
