package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ReneeB2022/test1/internal/data"
	"github.com/ReneeB2022/test1/internal/validator"
)

func (a *applicationDependencies) createstudentHandler(w http.ResponseWriter,
	r *http.Request) {
	// create a struct to hold a Department
	// we use struct tags[``] to make the names display in lowercase
	var incomingData struct {
		Fname      string  `json:"Fname"` // the student data
		Lname      string  `json:"Lname"`
		Gender     string  `json:"gender"`
		Age        int64   `json:"age"`
		District   int64   `json:"district_id"`
		Program_id int64   `json:"program"`
		GPA        float64 `json:"gpa"`
	}
	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.errorResponseJSON(w, r, http.StatusBadRequest, err.Error())
		return
	}
	stu := &data.Student{
		Fname:      incomingData.Fname,
		Lname:      incomingData.Lname,
		Gender:     incomingData.Gender,
		Age:        incomingData.Age,
		District:   incomingData.District,
		Program_id: incomingData.Program_id,
		GPA:        incomingData.GPA,
	}
	// Initialize a Validator instance
	v := validator.New()

	data.Validatestudent(v, stu)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors) // implemented later
		return
	}

	err = a.studentModel.Insert(stu)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/students/%d", stu.Student_id))

	// Send a JSON response with 201 (new resource created) status code
	data := envelope{
		"Student": stu,
	}
	err = a.writeJSON(w, http.StatusCreated, data, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

}
func (a *applicationDependencies) displaystudentHandler(
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
	stu, err := a.studentModel.Get(id)
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
		"Student": stu,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

}
func (a *applicationDependencies) updatestudentHandler(
	w http.ResponseWriter,
	r *http.Request) {
	// Get the id from the URL
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	// Call Get() to retrieve the Department with the specified id
	stu, err := a.studentModel.Get(id)
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
		Fname      string   `json:"Fname"` // the student data
		Lname      string   `json:"Lname"`
		Gender     string   `json:"gender"`
		Age        *int64   `json:"age"`
		District   *int64   `json:"district_id"`
		Program_id *int64   `json:"program"`
		GPA        *float64 `json:"gpa"`
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
		stu.Fname = incomingData.Fname
	}
	// if incomingData.Department_code is nil, no update was provided
	if incomingData.Lname != "" {
		stu.Lname = incomingData.Lname
	}
	if incomingData.Gender != "" {
		stu.Gender = incomingData.Gender
	}
	if incomingData.Age != nil {
		stu.Age = *incomingData.Age
	}
	if incomingData.District != nil {
		stu.District = *incomingData.District
	}
	if incomingData.Program_id != nil {
		stu.Program_id = *incomingData.Program_id
	}
	if incomingData.GPA != nil {
		stu.GPA = *incomingData.GPA
	}

	// Before we write the updates to the DB let's validate
	v := validator.New()
	data.Validatestudent(v, stu)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}
	// perform the update
	err = a.studentModel.Update(stu)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	data := envelope{
		"Student": stu,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}
func (a *applicationDependencies) deletestudentHandler(
	w http.ResponseWriter,
	r *http.Request) {

	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	err = a.studentModel.Delete(id)

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
		"message": "Student successfully deleted",
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}

}

func (a *applicationDependencies) liststudentHandler(
	w http.ResponseWriter,
	r *http.Request) {

	var queryParametersData struct {
		Fname    string
		Lname    string
		District int64
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

	queryParametersData.District = int64(a.getSingleIntegerParameter(
		queryParameters,
		"District",
		0,
		v,
	))

	queryParametersData.Filters.Page = a.getSingleIntegerParameter(
		queryParameters, "page", 1, v)
	queryParametersData.Filters.PageSize = a.getSingleIntegerParameter(
		queryParameters, "page_size", 10, v)

	queryParametersData.Filters.Sort = a.getSingleQueryParameter(
		queryParameters, "sort", "student_id")

	queryParametersData.Filters.SortSafeList = []string{"student_id", "Fname", "District",
		"-student_id", "-Fname", "-District"}

	stus, metadata, err := a.studentModel.GetAll(queryParametersData.Fname,
		queryParametersData.Lname, queryParametersData.District, queryParametersData.Filters)

	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{
		"Student":   stus,
		"@metadata": metadata,
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
