package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ReneeB2022/test1/internal/data"
	"github.com/ReneeB2022/test1/internal/validator"
)

func (a *applicationDependencies) createAdvisoryHandler(w http.ResponseWriter,
	r *http.Request) {
	// create a struct to hold a Advisory
	// we use struct tags[``] to make the names display in lowercase
	var incomingData struct {
		Advisor_id int64 `json:"advisor"`
		Student_id int64 `json:"student"` // the Advisory data
	}
	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.errorResponseJSON(w, r, http.StatusBadRequest, err.Error())
		return
	}
	adv := &data.Advisory{
		Advisor_id: incomingData.Advisor_id,
		Student_id: incomingData.Student_id,
	}
	// Initialize a Validator instance
	v := validator.New()

	data.ValidateAdvisors(v, adv)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors) // implemented later
		return
	}

	err = a.advisoryModel.Insert(adv)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/Advisorys/%d", adv.ADSTU_id))

	// Send a JSON response with 201 (new resource created) status code
	data := envelope{
		"Advisory": adv,
	}
	err = a.writeJSON(w, http.StatusCreated, data, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

}
func (a *applicationDependencies) displayadvisoryHandler(
	w http.ResponseWriter,
	r *http.Request) {

	// Get the id from the URL /v1/Advisorys/:id so that we
	// can use it to query teh Advisorys table. We will
	// implement the readIDParam() function later
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}
	// Call Get() to retrieve the Advisory with the specified id
	adv, err := a.advisoryModel.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}
	// display the Advisory
	data := envelope{
		"Advisory": adv,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

}
func (a *applicationDependencies) updateadvisorHandler(
	w http.ResponseWriter,
	r *http.Request) {
	// Get the id from the URL
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	// Call Get() to retrieve the Advisory with the specified id
	adv, err := a.advisoryModel.Get(id)
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
		Advisor_id *int64 `json:"advisor"`
		Student_id *int64 `json:"student"`
	}

	// perform the decoding
	err = a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}
	// We need to now check the fields to see which ones need updating
	// if incomingData.Advisory_name is nil, no update was provided
	if incomingData.Advisor_id != nil {
		adv.Advisor_id = *incomingData.Advisor_id
	}
	// if incomingData.Advisory_code is nil, no update was provided
	if incomingData.Student_id != nil {
		adv.Student_id = *incomingData.Student_id
	}

	// Before we write the updates to the DB let's validate
	v := validator.New()
	data.ValidateAdvisors(v, adv)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}
	// perform the update
	err = a.advisoryModel.Update(adv)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	data := envelope{
		"Advisory": adv,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}
func (a *applicationDependencies) deleteadvisorHandler(
	w http.ResponseWriter,
	r *http.Request) {

	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	err = a.advisoryModel.Delete(id)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	// display the Advisory
	data := envelope{
		"message": "Advisor successfully deleted",
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}

}

func (a *applicationDependencies) listadvisorHandler(
	w http.ResponseWriter,
	r *http.Request) {

	var queryParametersData struct {
		Advisor_id int64
		data.Filters
	}

	queryParameters := r.URL.Query()

	v := validator.New()
	queryParametersData.Advisor_id = int64(a.getSingleIntegerParameter(
		queryParameters,
		"Advisor",
		0,
		v,
	))

	queryParametersData.Filters.Page = a.getSingleIntegerParameter(
		queryParameters, "page", 1, v)
	queryParametersData.Filters.PageSize = a.getSingleIntegerParameter(
		queryParameters, "page_size", 10, v)

	queryParametersData.Filters.Sort = a.getSingleQueryParameter(
		queryParameters, "sort", "ADSTU_id")

	queryParametersData.Filters.SortSafeList = []string{"ADSTU_id", "advisor_id",
		"-ADSTU_id", "-advisor_id"}

	advs, metadata, err := a.advisoryModel.GetAll(queryParametersData.Advisor_id,
		queryParametersData.Filters)

	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{
		"Advisorys": advs,
		"@metadata": metadata,
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
