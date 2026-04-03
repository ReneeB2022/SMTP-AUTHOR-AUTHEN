package main

import (
	"expvar"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (a *applicationDependencies) routes() http.Handler {

	const datasize = 1024
	// setup a new router
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(a.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(a.methodNotAllowedResponse)

	// setup routes:

	//user
	router.HandlerFunc(http.MethodPost, "/v1/users", a.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", a.activateUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", a.createAuthenticationTokenHandler)

	//departments
	router.HandlerFunc(http.MethodGet, "/v1/printinfo", a.requirePermission("university:read", a.printinfo))
	router.HandlerFunc(http.MethodGet, "/v1/departments", a.requirePermission("university:read", a.listdeptHandler))
	router.HandlerFunc(http.MethodPost, "/v1/departments", a.requirePermission("university:write", a.createdepartmentHandler))
	router.HandlerFunc(http.MethodGet, "/v1/departments/:id", a.requirePermission("university:read", a.displaydeptHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/departments/:id", a.requirePermission("university:write", a.updatedeptHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/departments/:id", a.requirePermission("university:write", a.deletedeptHandler))

	//students
	router.HandlerFunc(http.MethodGet, "/v1/students", a.requirePermission("university:read", a.liststudentHandler))
	router.HandlerFunc(http.MethodPost, "/v1/students", a.requirePermission("university:write", a.createstudentHandler))
	router.HandlerFunc(http.MethodGet, "/v1/students/:id", a.requirePermission("university:read", a.displaystudentHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/students/:id", a.requirePermission("university:write", a.updatestudentHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/students/:id", a.requirePermission("university:write", a.deletestudentHandler))

	//Roles
	router.HandlerFunc(http.MethodGet, "/v1/roles", a.requirePermission("university:read", a.listroleHandler))
	router.HandlerFunc(http.MethodPost, "/v1/roles", a.requirePermission("university:write", a.createroleHandler))
	router.HandlerFunc(http.MethodGet, "/v1/roles/:id", a.requirePermission("university:read", a.displayroleHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/roles/:id", a.requirePermission("university:write", a.updateroleHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/roles/:id", a.requirePermission("university:write", a.deleteroleHandler))

	//Staff
	router.HandlerFunc(http.MethodGet, "/v1/staff", a.requirePermission("university:read", a.liststaffHandler))
	router.HandlerFunc(http.MethodPost, "/v1/staff", a.requirePermission("university:write", a.createstaffHandler))
	router.HandlerFunc(http.MethodGet, "/v1/staff/:id", a.requirePermission("university:read", a.displaystaffHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/staff/:id", a.requirePermission("university:write", a.updatestaffHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/staff/:id", a.requirePermission("university:write", a.deletestaffHandler))

	//District
	router.HandlerFunc(http.MethodGet, "/v1/district", a.requirePermission("university:read", a.listdistHandler))
	router.HandlerFunc(http.MethodPost, "/v1/district", a.requirePermission("university:write", a.createdistHandler))
	router.HandlerFunc(http.MethodGet, "/v1/district/:id", a.requirePermission("university:read", a.displaydistHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/district/:id", a.requirePermission("university:write", a.updatedistHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/district/:id", a.requirePermission("university:write", a.deletedistHandler))

	//Programs
	router.HandlerFunc(http.MethodGet, "/v1/programs", a.requirePermission("university:read", a.listprogramHandler))
	router.HandlerFunc(http.MethodPost, "/v1/programs", a.requirePermission("university:write", a.createprogramHandler))
	router.HandlerFunc(http.MethodGet, "/v1/programs/:id", a.requirePermission("university:read", a.displayprogramHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/programs/:id", a.requirePermission("university:write", a.updateprogramHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/programs/:id", a.requirePermission("university:write", a.deleteprogramHandler))

	//enrollments
	router.HandlerFunc(http.MethodGet, "/v1/enrollments", a.requirePermission("university:read", a.listenrollHandler))
	router.HandlerFunc(http.MethodPost, "/v1/enrollments", a.requirePermission("university:write", a.createenrollrec))
	router.HandlerFunc(http.MethodGet, "/v1/enrollments/:id", a.requirePermission("university:read", a.displayenrollHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/enrollments/:id", a.requirePermission("university:write", a.updateenrollHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/enrollments/:id", a.requirePermission("university:write", a.deleteenrollHandler))

	//Advisors
	router.HandlerFunc(http.MethodGet, "/v1/advisors", a.requirePermission("university:read", a.listadvisorHandler))
	router.HandlerFunc(http.MethodPost, "/v1/advisors", a.requirePermission("university:write", a.createAdvisoryHandler))
	router.HandlerFunc(http.MethodGet, "/v1/advisors/:id", a.requirePermission("university:read", a.displayadvisoryHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/advisors/:id", a.requirePermission("university:write", a.updateadvisorHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/advisors/:id", a.requirePermission("university:write", a.deleteadvisorHandler))

	//Courses
	router.HandlerFunc(http.MethodGet, "/v1/courses", a.requirePermission("university:read", a.listcourseHandler))
	router.HandlerFunc(http.MethodPost, "/v1/courses", a.requirePermission("university:write", a.createcourseHandler))
	router.HandlerFunc(http.MethodGet, "/v1/courses/:id", a.requirePermission("university:read", a.displaycourseHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/courses/:id", a.requirePermission("university:write", a.updatecourseHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/courses/:id", a.requirePermission("university:write", a.deletecourseHandler))

	//Grades
	router.HandlerFunc(http.MethodGet, "/v1/grades", a.requirePermission("university:read", a.listGradeHandler))
	router.HandlerFunc(http.MethodPost, "/v1/grades", a.requirePermission("university:write", a.createGradeHandler))
	router.HandlerFunc(http.MethodGet, "/v1/grades/:id", a.requirePermission("university:read", a.displayGradeHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/grades/:id", a.requirePermission("university:write", a.updateGradeHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/grades/:id", a.requirePermission("university:write", a.deleteGradeHandler))

	//sections
	router.HandlerFunc(http.MethodGet, "/v1/sections", a.requirePermission("university:read", a.listSectionHandler))
	router.HandlerFunc(http.MethodPost, "/v1/sections", a.requirePermission("university:write", a.createSectionHandler))
	router.HandlerFunc(http.MethodGet, "/v1/sections/:id", a.requirePermission("university:read", a.displaySectionHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/sections/:id", a.requirePermission("university:write", a.updateSectionHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/sections/:id", a.requirePermission("university:write", a.deleteSectionHandler))

	//metrics
	router.Handler(http.MethodGet, "/v1/observability/database/metrics", expvar.Handler())

	return a.metrics(a.recoverPanic(a.enableCORS(a.rateLimit(a.authenticate(LoggingMiddleware(GzipMiddleware(datasize, router)))))))

}
