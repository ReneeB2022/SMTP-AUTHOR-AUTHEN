package main

import (
	"fmt"
	"net/http"
)

// We set the WWW-Authenticate header to give a hint to the user as
// to what they need to provide. Don't want to leave them guessing
func (a *applicationDependencies) invalidAuthenticationTokenResponse(w http.ResponseWriter,
	r *http.Request) {

	w.Header().Set("WWW-Authenticate", "Bearer")

	message := "invalid or missing authentication token"
	a.errorResponseJSON(w, r, http.StatusUnauthorized, message)

}

// Note: 401 is Unauthorized (anonymous) and 403 is Forbidden (authenticated
// but not activated or don't have the right privilege)

func (a *applicationDependencies) authenticationRequiredResponse(w http.ResponseWriter, r *http.Request) {
	message := "you must be authenticated to access this resource"
	a.errorResponseJSON(w, r, http.StatusUnauthorized, message)
}

func (a *applicationDependencies) inactiveAccountResponse(w http.ResponseWriter, r *http.Request) {
	message := "your user account must be activated to access this resource"
	a.errorResponseJSON(w, r, http.StatusForbidden, message)
}

func (a *applicationDependencies) notPermittedResponse(w http.ResponseWriter,
	r *http.Request) {
	message := "your user account doesn't have the necessary permissions to access this resource"

	a.errorResponseJSON(w, r, http.StatusForbidden, message)
}

func (a *applicationDependencies) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {

	message := "rate limit exceeded"
	a.errorResponseJSON(w, r, http.StatusTooManyRequests, message)
}

func (a *applicationDependencies) logError(r *http.Request, err error) {

	method := r.Method
	uri := r.URL.RequestURI()
	a.logger.Error(err.Error(), "method", method, "uri", uri)

}
func (a *applicationDependencies) errorResponseJSON(w http.ResponseWriter,
	r *http.Request,
	status int,
	message any) {

	errorData := envelope{"error": message}
	err := a.writeJSON(w, status, errorData, nil)
	if err != nil {
		a.logError(r, err)
		w.WriteHeader(500)
	}
}
func (a *applicationDependencies) serverErrorResponse(w http.ResponseWriter,
	r *http.Request,
	err error) {

	// first thing is to log error message
	a.logError(r, err)
	// prepare a response to send to the client
	message := "the server encountered a problem and could not process your request"
	a.errorResponseJSON(w, r, http.StatusInternalServerError, message)
}
func (a *applicationDependencies) notFoundResponse(w http.ResponseWriter,
	r *http.Request) {

	// we only log server errors, not client errors
	// prepare a response to send to the client
	message := "the requested resource could not be found"
	a.errorResponseJSON(w, r, http.StatusNotFound, message)
}
func (a *applicationDependencies) methodNotAllowedResponse(
	w http.ResponseWriter,
	r *http.Request) {

	// we only log server errors, not client errors
	// prepare a formatted response to send to the client
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)

	a.errorResponseJSON(w, r, http.StatusMethodNotAllowed, message)
}

func (a *applicationDependencies) failedValidationResponse(w http.ResponseWriter, r *http.Request,
	errors map[string]string) {
	a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, errors)
}
func (a *applicationDependencies) badRequestResponse(w http.ResponseWriter,
	r *http.Request,
	err error) {

	a.errorResponseJSON(w, r, http.StatusBadRequest, err.Error())
}

func (a *applicationDependencies) editConflictResponse(w http.ResponseWriter,
	r *http.Request) {

	message := "unable to update the record due to an edit conflict, please try again"
	a.errorResponseJSON(w, r, http.StatusConflict, message)

}
