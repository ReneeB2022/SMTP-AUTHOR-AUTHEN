package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ReneeB2022/test1/internal/data"
	"github.com/ReneeB2022/test1/internal/validator"
)

func (a *applicationDependencies) createcourseHandler(w http.ResponseWriter,
	r *http.Request) {
	// create a struct to hold a Department
	// we use struct tags[``] to make the names display in lowercase
	var incomingData struct {
		Course_name   string  `json:"Course"` // the Course data
		Course_code   string  `json:"Code"`
		Credits       int64   `json:"credits"`
		Descriptions  string  `json:"description"`
		Prerequisites string  `json:"prereq"`
		Fee           float64 `json:"fee"`
		Dept_id       int64   `json:"DeptId"`
	}
	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.errorResponseJSON(w, r, http.StatusBadRequest, err.Error())
		return
	}
	cor := &data.Course{
		Course_name:   incomingData.Course_name,
		Course_code:   incomingData.Course_code,
		Credits:       incomingData.Credits,
		Descriptions:  incomingData.Descriptions,
		Prerequisites: incomingData.Prerequisites,
		Fee:           incomingData.Fee,
		Dept_id:       incomingData.Dept_id,
	}
	// Initialize a Validator instance
	v := validator.New()

	data.ValidateCourse(v, cor)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors) // implemented later
		return
	}

	err = a.courseModel.Insert(cor)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/courses/%d", cor.Course_id))

	// Send a JSON response with 201 (new resource created) status code
	data := envelope{
		"Course": cor,
	}
	err = a.writeJSON(w, http.StatusCreated, data, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

}
func (a *applicationDependencies) displaycourseHandler(
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
	cor, err := a.courseModel.Get(id)
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
		"Course": cor,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

}
func (a *applicationDependencies) updatecourseHandler(
	w http.ResponseWriter,
	r *http.Request) {
	// Get the id from the URL
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	// Call Get() to retrieve the Department with the specified id
	cor, err := a.courseModel.Get(id)
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
		Course_name   string   `json:"Course"` // the Course data
		Course_code   string   `json:"Code"`
		Credits       *int64   `json:"credits"`
		Descriptions  string   `json:"description"`
		Prerequisites string   `json:"prereq"`
		Fee           *float64 `json:"fee"`
		Dept_id       *int64   `json:"DeptId"`
	}

	// perform the decoding
	err = a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}
	// We need to now check the fields to see which ones need updating
	// if incomingData.Department_name is nil, no update was provided
	if incomingData.Course_name != "" {
		cor.Course_name = incomingData.Course_name
	}
	// if incomingData.Department_code is nil, no update was provided
	if incomingData.Course_code != "" {
		cor.Course_code = incomingData.Course_code
	}
	if incomingData.Credits != nil {
		cor.Credits = *incomingData.Credits
	}
	if incomingData.Descriptions != "" {
		cor.Descriptions = incomingData.Descriptions
	}
	if incomingData.Prerequisites != "" {
		cor.Prerequisites = incomingData.Prerequisites
	}
	if incomingData.Fee != nil {
		cor.Fee = *incomingData.Fee
	}
	if incomingData.Dept_id != nil {
		cor.Dept_id = *incomingData.Dept_id
	}
	// Before we write the updates to the DB let's validate
	v := validator.New()
	data.ValidateCourse(v, cor)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}
	// perform the update
	err = a.courseModel.Update(cor)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	data := envelope{
		"Course": cor,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}
func (a *applicationDependencies) deletecourseHandler(
	w http.ResponseWriter,
	r *http.Request) {

	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	err = a.courseModel.Delete(id)

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
		"message": "Course successfully deleted",
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}

}

func (a *applicationDependencies) listcourseHandler(
	w http.ResponseWriter,
	r *http.Request) {

	var queryParametersData struct {
		Course_name string
		course_code string
		data.Filters
	}

	queryParameters := r.URL.Query()

	queryParametersData.Course_name = a.getSingleQueryParameter(
		queryParameters,
		"corname",
		"")

	queryParametersData.course_code = a.getSingleQueryParameter(
		queryParameters,
		"corcode",
		"")

	v := validator.New()

	queryParametersData.Filters.Page = a.getSingleIntegerParameter(
		queryParameters, "page", 1, v)
	queryParametersData.Filters.PageSize = a.getSingleIntegerParameter(
		queryParameters, "page_size", 10, v)

	queryParametersData.Filters.Sort = a.getSingleQueryParameter(
		queryParameters, "sort", "course_id")

	queryParametersData.Filters.SortSafeList = []string{"course_id", "course_name", "course_code",
		"-course_id", "-course_name", "-course_code"}

	cor, metadata, err := a.courseModel.GetAll(queryParametersData.Course_name,
		queryParametersData.course_code, queryParametersData.Filters)

	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{
		"Course":    cor,
		"@metadata": metadata,
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
