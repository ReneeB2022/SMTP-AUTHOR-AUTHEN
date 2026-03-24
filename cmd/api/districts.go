package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ReneeB2022/test1/internal/data"
	"github.com/ReneeB2022/test1/internal/validator"
)

func (a *applicationDependencies) createdistHandler(w http.ResponseWriter,
	r *http.Request) {
	// create a struct to hold a Department
	// we use struct tags[``] to make the names display in lowercase
	var incomingData struct {
		District_name string `json:"district"`
	}
	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.errorResponseJSON(w, r, http.StatusBadRequest, err.Error())
		return
	}
	dist := &data.District{
		District_name: incomingData.District_name,
	}
	// Initialize a Validator instance
	v := validator.New()

	data.ValidateDistrict(v, dist)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors) // implemented later
		return
	}

	err = a.districtModel.Insert(dist)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/districts/%d", dist.District_id))

	// Send a JSON response with 201 (new resource created) status code
	data := envelope{
		"district": dist,
	}
	err = a.writeJSON(w, http.StatusCreated, data, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

}
func (a *applicationDependencies) displaydistHandler(
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
	dist, err := a.districtModel.Get(id)
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
		"district": dist,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

}
func (a *applicationDependencies) updatedistHandler(
	w http.ResponseWriter,
	r *http.Request) {
	// Get the id from the URL
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	// Call Get() to retrieve the Department with the specified id
	dist, err := a.districtModel.Get(id)
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
		District_name string `json:"district"`
	}

	// perform the decoding
	err = a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}
	// We need to now check the fields to see which ones need updating
	// if incomingData.Department_name is nil, no update was provided
	if incomingData.District_name != "" {
		dist.District_name = incomingData.District_name
	}

	// Before we write the updates to the DB let's validate
	v := validator.New()
	data.ValidateDistrict(v, dist)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}
	// perform the update
	err = a.districtModel.Update(dist)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	data := envelope{
		"district": dist,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}
func (a *applicationDependencies) deletedistHandler(
	w http.ResponseWriter,
	r *http.Request) {

	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	err = a.districtModel.Delete(id)

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
		"message": "District successfully deleted",
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}

}

func (a *applicationDependencies) listdistHandler(
	w http.ResponseWriter,
	r *http.Request) {

	var queryParametersData struct {
		District_name string
		data.Filters
	}

	queryParameters := r.URL.Query()

	queryParametersData.District_name = a.getSingleQueryParameter(
		queryParameters,
		"dis",
		"")

	v := validator.New()

	queryParametersData.Filters.Page = a.getSingleIntegerParameter(
		queryParameters, "page", 1, v)
	queryParametersData.Filters.PageSize = a.getSingleIntegerParameter(
		queryParameters, "page_size", 10, v)

	queryParametersData.Filters.Sort = a.getSingleQueryParameter(
		queryParameters, "sort", "district_id")

	queryParametersData.Filters.SortSafeList = []string{"district_id", "district_name",
		"-district_id", "-district_name"}

	dist, metadata, err := a.rolemodel.GetAll(queryParametersData.District_name, queryParametersData.Filters)

	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{
		"district":  dist,
		"@metadata": metadata,
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
