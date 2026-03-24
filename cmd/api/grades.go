package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ReneeB2022/test1/internal/data"
	"github.com/ReneeB2022/test1/internal/validator"
)

func (a *applicationDependencies) createGradeHandler(w http.ResponseWriter,
	r *http.Request) {
	// create a struct to hold a Department
	// we use struct tags[``] to make the names display in lowercase
	var incomingData struct {
		Grades      string  `json:"Letter"`
		Grade_value float64 `json:"Value"`
	}
	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.errorResponseJSON(w, r, http.StatusBadRequest, err.Error())
		return
	}
	grade := &data.Grade{
		Grades:      incomingData.Grades,
		Grade_value: incomingData.Grade_value,
	}
	// Initialize a Validator instance
	v := validator.New()

	data.ValidateGrades(v, grade)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors) // implemented later
		return
	}

	err = a.grademodel.Insert(grade)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/Grades/%d", grade.Grade_id))

	// Send a JSON response with 201 (new resource created) status code
	data := envelope{
		"Grade": grade,
	}
	err = a.writeJSON(w, http.StatusCreated, data, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

}
func (a *applicationDependencies) displayGradeHandler(
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
	grade, err := a.grademodel.Get(id)
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
		"Grade": grade,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

}
func (a *applicationDependencies) updateGradeHandler(
	w http.ResponseWriter,
	r *http.Request) {
	// Get the id from the URL
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	// Call Get() to retrieve the Department with the specified id
	grade, err := a.grademodel.Get(id)
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
		Grades      string   `json:"Letter"`
		Grade_value *float64 `json:"Value"`
	}

	// perform the decoding
	err = a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}
	// We need to now check the fields to see which ones need updating
	// if incomingData.Department_name is nil, no update was gradevided
	if incomingData.Grades != "" {
		grade.Grades = incomingData.Grades
	}
	if incomingData.Grade_value != nil {
		grade.Grade_value = *incomingData.Grade_value
	}
	// if incomingData.Department_code is nil, no update was gradevided

	// Before we write the updates to the DB let's validate
	v := validator.New()
	data.ValidateGrades(v, grade)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}
	// perform the update
	err = a.grademodel.Update(grade)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	data := envelope{
		"Grade": grade,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}
func (a *applicationDependencies) deleteGradeHandler(
	w http.ResponseWriter,
	r *http.Request) {

	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	err = a.grademodel.Delete(id)

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
		"message": "Grade successfully deleted",
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}

}

func (a *applicationDependencies) listGradeHandler(
	w http.ResponseWriter,
	r *http.Request) {

	var queryParametersData struct {
		Grades string
		data.Filters
	}

	queryParameters := r.URL.Query()

	queryParametersData.Grades = a.getSingleQueryParameter(
		queryParameters,
		"grade",
		"")

	v := validator.New()

	queryParametersData.Filters.Page = a.getSingleIntegerParameter(
		queryParameters, "page", 1, v)
	queryParametersData.Filters.PageSize = a.getSingleIntegerParameter(
		queryParameters, "page_size", 10, v)

	queryParametersData.Filters.Sort = a.getSingleQueryParameter(
		queryParameters, "sort", "Grade_id")

	queryParametersData.Filters.SortSafeList = []string{"Grade_id", "Grade_code",
		"-Grade_id", "-Grade_code"}

	grade, metadata, err := a.grademodel.GetAll(queryParametersData.Grades, queryParametersData.Filters)

	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{
		"Grades":    grade,
		"@metadata": metadata,
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
