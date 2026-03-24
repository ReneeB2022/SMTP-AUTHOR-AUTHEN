package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ReneeB2022/test1/internal/data"
	"github.com/ReneeB2022/test1/internal/validator"
)

func (a *applicationDependencies) createdepartmentHandler(w http.ResponseWriter,
	r *http.Request) {
	// create a struct to hold a Department
	// we use struct tags[``] to make the names display in lowercase
	var incomingData struct {
		Department_name string `json:"Deptname"`
		Department_code string `json:"Deptcode"`
	}
	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.errorResponseJSON(w, r, http.StatusBadRequest, err.Error())
		return
	}
	dept := &data.Department{
		Department_name: incomingData.Department_name,
		Department_code: incomingData.Department_code,
	}
	// Initialize a Validator instance
	v := validator.New()

	data.Validatedepartment(v, dept)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors) // implemented later
		return
	}

	err = a.departmentModel.Insert(dept)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/departments/%d", dept.Department_id))

	// Send a JSON response with 201 (new resource created) status code
	data := envelope{
		"department": dept,
	}
	err = a.writeJSON(w, http.StatusCreated, data, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

}
func (a *applicationDependencies) displaydeptHandler(
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
	dept, err := a.departmentModel.Get(id)
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
		"department": dept,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

}
func (a *applicationDependencies) updatedeptHandler(
	w http.ResponseWriter,
	r *http.Request) {
	// Get the id from the URL
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	// Call Get() to retrieve the Department with the specified id
	dept, err := a.departmentModel.Get(id)
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
		Department_name string `json:"Deptname"`
		Department_code string `json:"Deptcode"`
	}

	// perform the decoding
	err = a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}
	// We need to now check the fields to see which ones need updating
	// if incomingData.Department_name is nil, no update was provided
	if incomingData.Department_name != "" {
		dept.Department_name = incomingData.Department_name
	}
	// if incomingData.Department_code is nil, no update was provided
	if incomingData.Department_code != "" {
		dept.Department_code = incomingData.Department_code
	}

	// Before we write the updates to the DB let's validate
	v := validator.New()
	data.Validatedepartment(v, dept)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}
	// perform the update
	err = a.departmentModel.Update(dept)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	data := envelope{
		"department": dept,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}
func (a *applicationDependencies) deletedeptHandler(
	w http.ResponseWriter,
	r *http.Request) {

	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	err = a.departmentModel.Delete(id)

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
		"message": "Department successfully deleted",
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}

}

func (a *applicationDependencies) listdeptHandler(
	w http.ResponseWriter,
	r *http.Request) {

	var queryParametersData struct {
		Department_name string
		Department_code string
		data.Filters
	}

	queryParameters := r.URL.Query()

	queryParametersData.Department_name = a.getSingleQueryParameter(
		queryParameters,
		"deptname",
		"")

	queryParametersData.Department_code = a.getSingleQueryParameter(
		queryParameters,
		"deptcode",
		"")

	v := validator.New()

	queryParametersData.Filters.Page = a.getSingleIntegerParameter(
		queryParameters, "page", 1, v)
	queryParametersData.Filters.PageSize = a.getSingleIntegerParameter(
		queryParameters, "page_size", 10, v)

	queryParametersData.Filters.Sort = a.getSingleQueryParameter(
		queryParameters, "sort", "department_id")

	queryParametersData.Filters.SortSafeList = []string{"department_id", "department_name",
		"-department_id", "-department_name"}

	depts, metadata, err := a.departmentModel.GetAll(queryParametersData.Department_name,
		queryParametersData.Department_code, queryParametersData.Filters)

	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{
		"departments": depts,
		"@metadata":   metadata,
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
