package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ReneeB2022/test1/internal/data"
	"github.com/ReneeB2022/test1/internal/validator"
)

func (a *applicationDependencies) createenrollrec(w http.ResponseWriter,
	r *http.Request) {
	// create a struct to hold a Enrollment
	// we use struct tags[``] to make the names display in lowercase
	var incomingData struct {
		Student_id int64     `json:"student"` // the Enrollment data
		Section_id int64     `json:"section"`
		Grade_id   int64     `json:"grade"`
		Date       time.Time `json:"enrolled"`
	}
	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.errorResponseJSON(w, r, http.StatusBadRequest, err.Error())
		return
	}
	enroll := &data.Enrollment{
		Student_id: incomingData.Student_id,
		Section_id: incomingData.Section_id,
		Grade_id:   incomingData.Grade_id,
		Date:       incomingData.Date,
	}
	// Initialize a Validator instance
	v := validator.New()

	data.ValidateEnrollment(v, enroll)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors) // implemented later
		return
	}

	err = a.enrollmentModel.Insert(enroll)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/Enrollments/%d", enroll.Enrollment_id))

	// Send a JSON response with 201 (new resource created) status code
	data := envelope{
		"Enrollment": enroll,
	}
	err = a.writeJSON(w, http.StatusCreated, data, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

}
func (a *applicationDependencies) displayenrollHandler(
	w http.ResponseWriter,
	r *http.Request) {

	// Get the id from the URL /v1/Enrollments/:id so that we
	// can use it to query teh Enrollments table. We will
	// implement the readIDParam() function later
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}
	// Call Get() to retrieve the Enrollment with the specified id
	enroll, err := a.enrollmentModel.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}
	// display the Enrollment
	data := envelope{
		"Enrollment": enroll,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

}
func (a *applicationDependencies) updateenrollHandler(
	w http.ResponseWriter,
	r *http.Request) {
	// Get the id from the URL
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	// Call Get() to retrieve the Enrollment with the specified id
	enroll, err := a.enrollmentModel.Get(id)
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
		Grade_id *int64 `json:"grade"`
	}

	// perform the decoding
	err = a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	if incomingData.Grade_id != nil {
		enroll.Grade_id = *incomingData.Grade_id
	}
	// Before we write the updates to the DB let's validate
	v := validator.New()
	data.ValidateEnrollment(v, enroll)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}
	// perform the update
	err = a.enrollmentModel.Update(enroll)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	data := envelope{
		"Enrollment": enroll,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}
func (a *applicationDependencies) deleteenrollHandler(
	w http.ResponseWriter,
	r *http.Request) {

	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	err = a.enrollmentModel.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	// display the Enrollment
	data := envelope{
		"message": "Enrollment successfully deleted",
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}

}

func (a *applicationDependencies) listenrollHandler(
	w http.ResponseWriter,
	r *http.Request) {

	var queryParametersData struct {
		Section_id int64
		Grade_id   int64
		data.Filters
	}

	queryParameters := r.URL.Query()

	v := validator.New()

	queryParametersData.Section_id = int64(a.getSingleIntegerParameter(
		queryParameters,
		"Section",
		0,
		v,
	))
	queryParametersData.Grade_id = int64(a.getSingleIntegerParameter(
		queryParameters,
		"Grade",
		0,
		v,
	))
	queryParametersData.Filters.Page = a.getSingleIntegerParameter(
		queryParameters, "page", 1, v)
	queryParametersData.Filters.PageSize = a.getSingleIntegerParameter(
		queryParameters, "page_size", 10, v)

	queryParametersData.Filters.Sort = a.getSingleQueryParameter(
		queryParameters, "sort", "enrollment_id")

	queryParametersData.Filters.SortSafeList = []string{"enrollment_id", "section_id",
		"-enrollment_id", "-section_id"}

	enroll, metadata, err := a.enrollmentModel.GetAll(queryParametersData.Section_id,
		queryParametersData.Grade_id, queryParametersData.Filters)

	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{
		"Enrollments": enroll,
		"@metadata":   metadata,
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
