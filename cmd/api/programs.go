package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ReneeB2022/test1/internal/data"
	"github.com/ReneeB2022/test1/internal/validator"
)

func (a *applicationDependencies) createprogramHandler(w http.ResponseWriter,
	r *http.Request) {
	// create a struct to hold a Department
	// we use struct tags[``] to make the names display in lowercase
	var incomingData struct {
		Programs      string `json:"program"`
		Program_code  string `json:"Procode"`
		Department_id int64  `json:"Deptid"`
	}
	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.errorResponseJSON(w, r, http.StatusBadRequest, err.Error())
		return
	}
	pro := &data.Program{
		Programs:      incomingData.Programs,
		Program_code:  incomingData.Program_code,
		Department_id: incomingData.Department_id,
	}
	// Initialize a Validator instance
	v := validator.New()

	data.Validateprograms(v, pro)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors) // implemented later
		return
	}

	err = a.programModel.Insert(pro)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/programs/%d", pro.Program_id))

	// Send a JSON response with 201 (new resource created) status code
	data := envelope{
		"Program": pro,
	}
	err = a.writeJSON(w, http.StatusCreated, data, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

}
func (a *applicationDependencies) displayprogramHandler(
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
	pro, err := a.programModel.Get(id)
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
		"Program": pro,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

}
func (a *applicationDependencies) updateprogramHandler(
	w http.ResponseWriter,
	r *http.Request) {
	// Get the id from the URL
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	// Call Get() to retrieve the Department with the specified id
	pro, err := a.programModel.Get(id)
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
		Programs      string `json:"program"`
		Program_code  string `json:"Procode"`
		Department_id *int64 `json:"Deptid"`
	}

	// perform the decoding
	err = a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}
	// We need to now check the fields to see which ones need updating
	// if incomingData.Department_name is nil, no update was provided
	if incomingData.Programs != "" {
		pro.Programs = incomingData.Programs
	}
	if incomingData.Program_code != "" {
		pro.Program_code = incomingData.Program_code
	}
	// if incomingData.Department_code is nil, no update was provided
	if incomingData.Department_id != nil {
		pro.Department_id = *incomingData.Department_id
	}

	// Before we write the updates to the DB let's validate
	v := validator.New()
	data.Validateprograms(v, pro)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}
	// perform the update
	err = a.programModel.Update(pro)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	data := envelope{
		"Program": pro,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}
func (a *applicationDependencies) deleteprogramHandler(
	w http.ResponseWriter,
	r *http.Request) {

	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	err = a.programModel.Delete(id)

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
		"message": "Program successfully deleted",
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}

}

func (a *applicationDependencies) listprogramHandler(
	w http.ResponseWriter,
	r *http.Request) {

	var queryParametersData struct {
		Program_code  string
		Department_id int64
		data.Filters
	}

	queryParameters := r.URL.Query()

	queryParametersData.Program_code = a.getSingleQueryParameter(
		queryParameters,
		"procode",
		"")

	v := validator.New()

	queryParametersData.Department_id = int64(a.getSingleIntegerParameter(
		queryParameters,
		"Department",
		0,
		v,
	))

	queryParametersData.Filters.Page = a.getSingleIntegerParameter(
		queryParameters, "page", 1, v)
	queryParametersData.Filters.PageSize = a.getSingleIntegerParameter(
		queryParameters, "page_size", 10, v)

	queryParametersData.Filters.Sort = a.getSingleQueryParameter(
		queryParameters, "sort", "program_id")

	queryParametersData.Filters.SortSafeList = []string{"program_id", "program_code",
		"-program_id", "-program_code"}

	pro, metadata, err := a.programModel.GetAll(queryParametersData.Program_code,
		queryParametersData.Department_id, queryParametersData.Filters)

	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{
		"Programs":  pro,
		"@metadata": metadata,
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
