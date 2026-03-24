package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ReneeB2022/test1/internal/data"
	"github.com/ReneeB2022/test1/internal/validator"
)

func (a *applicationDependencies) createSectionHandler(w http.ResponseWriter,
	r *http.Request) {
	// create a struct to hold a Department
	// we use struct tags[``] to make the names display in lowercase
	var incomingData struct {
		Staff_id     int64     `json:"Lecturer"`
		Course_id    int64     `json:"Course_id"`
		Availability int64     `json:"available"`
		Classroom    string    `json:"Room"`
		Classday     string    `json:"Day"`
		Start        time.Time `json:"starts"`
		End          time.Time `json:"finish"`
		Semester     string    `json:"semester"`
	}
	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.errorResponseJSON(w, r, http.StatusBadRequest, err.Error())
		return
	}
	sec := &data.Section{
		Staff_id:     incomingData.Staff_id,
		Course_id:    incomingData.Course_id,
		Availability: incomingData.Availability,
		Classroom:    incomingData.Classroom,
		Classday:     incomingData.Classday,
		Start:        incomingData.Start,
		End:          incomingData.End,
		Semester:     incomingData.Semester,
	}
	// Initialize a Validator instance
	v := validator.New()

	data.ValidateSection(v, sec)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors) // implemented later
		return
	}

	err = a.sectionModel.Insert(sec)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/Sections/%d", sec.Section_id))

	// Send a JSON response with 201 (new resource created) status code
	data := envelope{
		"Section": sec,
	}
	err = a.writeJSON(w, http.StatusCreated, data, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

}
func (a *applicationDependencies) displaySectionHandler(
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
	sec, err := a.sectionModel.Get(id)
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
		"Section": sec,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

}
func (a *applicationDependencies) updateSectionHandler(
	w http.ResponseWriter,
	r *http.Request) {
	// Get the id from the URL
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	// Call Get() to retrieve the Department with the specified id
	sec, err := a.sectionModel.Get(id)
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
		Staff_id     *int64     `json:"Lecturer"`
		Course_id    *int64     `json:"Course_id"`
		Availability *int64     `json:"available"`
		Classroom    string     `json:"Room"`
		Classday     string     `json:"Day"`
		Start        *time.Time `json:"starts"`
		End          *time.Time `json:"finish"`
		Semester     string     `json:"semester"`
	}

	// perform the decoding
	err = a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}
	// We need to now check the fields to see which ones need updating
	// if incomingData. is nil, no update was provided
	if incomingData.Staff_id != nil {
		sec.Staff_id = *incomingData.Staff_id
	}
	// if incomingData. is nil, no update was provided
	if incomingData.Course_id != nil {
		sec.Course_id = *incomingData.Course_id
	}
	if incomingData.Availability != nil {
		sec.Availability = *incomingData.Availability
	}
	if incomingData.Classroom != "" {
		sec.Classroom = incomingData.Classroom
	}
	if incomingData.Classday != "" {
		sec.Classday = incomingData.Classday
	}
	if incomingData.Semester != "" {
		sec.Semester = incomingData.Semester
	}
	if incomingData.Start != nil {
		sec.Start = *incomingData.Start
	}
	if incomingData.End != nil {
		sec.End = *incomingData.End
	}
	// Before we write the updates to the DB let's validate
	v := validator.New()
	data.ValidateSection(v, sec)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}
	// perform the update
	err = a.sectionModel.Update(sec)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	data := envelope{
		"Section": sec,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}
func (a *applicationDependencies) deleteSectionHandler(
	w http.ResponseWriter,
	r *http.Request) {

	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	err = a.sectionModel.Delete(id)

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
		"message": "Section successfully deleted",
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}

}

func (a *applicationDependencies) listSectionHandler(
	w http.ResponseWriter,
	r *http.Request) {

	var queryParametersData struct {
		courseid int64
		data.Filters
	}

	queryParameters := r.URL.Query()

	v := validator.New()

	queryParametersData.courseid = int64(a.getSingleIntegerParameter(
		queryParameters,
		"Courseid",
		0,
		v,
	))

	queryParametersData.Filters.Page = a.getSingleIntegerParameter(
		queryParameters, "page", 1, v)
	queryParametersData.Filters.PageSize = a.getSingleIntegerParameter(
		queryParameters, "page_size", 10, v)

	queryParametersData.Filters.Sort = a.getSingleQueryParameter(
		queryParameters, "sort", "section_id")

	queryParametersData.Filters.SortSafeList = []string{"section_id", "course_id",
		"-section_id", "-course_id"}

	secs, metadata, err := a.sectionModel.GetAll(queryParametersData.courseid, queryParametersData.Filters)

	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{
		"Section":   secs,
		"@metadata": metadata,
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
