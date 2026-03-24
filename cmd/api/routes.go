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

	//departments
	router.HandlerFunc(http.MethodGet, "/v1/printinfo", a.printinfo)
	router.HandlerFunc(http.MethodGet, "/v1/departments", a.listdeptHandler)
	router.HandlerFunc(http.MethodPost, "/v1/departments", a.createdepartmentHandler)
	router.HandlerFunc(http.MethodGet, "/v1/departments/:id", a.displaydeptHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/departments/:id", a.updatedeptHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/departments/:id", a.deletedeptHandler)

	//students
	router.HandlerFunc(http.MethodGet, "/v1/students", a.liststudentHandler)
	router.HandlerFunc(http.MethodPost, "/v1/students", a.createstudentHandler)
	router.HandlerFunc(http.MethodGet, "/v1/students/:id", a.displaystudentHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/students/:id", a.updatestudentHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/students/:id", a.deletestudentHandler)

	//Roles
	router.HandlerFunc(http.MethodGet, "/v1/roles", a.listroleHandler)
	router.HandlerFunc(http.MethodPost, "/v1/roles", a.createroleHandler)
	router.HandlerFunc(http.MethodGet, "/v1/roles/:id", a.displayroleHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/roles/:id", a.updateroleHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/roles/:id", a.deleteroleHandler)

	//Staff
	router.HandlerFunc(http.MethodGet, "/v1/staff", a.liststaffHandler)
	router.HandlerFunc(http.MethodPost, "/v1/staff", a.createstaffHandler)
	router.HandlerFunc(http.MethodGet, "/v1/staff/:id", a.displaystaffHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/staff/:id", a.updatestaffHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/staff/:id", a.deletestaffHandler)

	//District
	router.HandlerFunc(http.MethodGet, "/v1/district", a.listdistHandler)
	router.HandlerFunc(http.MethodPost, "/v1/district", a.createdistHandler)
	router.HandlerFunc(http.MethodGet, "/v1/district/:id", a.displaydistHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/district/:id", a.updatedistHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/district/:id", a.deletedistHandler)

	//Programs
	router.HandlerFunc(http.MethodGet, "/v1/programs", a.listprogramHandler)
	router.HandlerFunc(http.MethodPost, "/v1/programs", a.createprogramHandler)
	router.HandlerFunc(http.MethodGet, "/v1/programs/:id", a.displayprogramHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/programs/:id", a.updateprogramHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/programs/:id", a.deleteprogramHandler)

	//enrollments
	router.HandlerFunc(http.MethodGet, "/v1/enrollments", a.listenrollHandler)
	router.HandlerFunc(http.MethodPost, "/v1/enrollments", a.createenrollrec)
	router.HandlerFunc(http.MethodGet, "/v1/enrollments/:id", a.displayenrollHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/enrollments/:id", a.updateenrollHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/enrollments/:id", a.deleteenrollHandler)

	//Advisors
	router.HandlerFunc(http.MethodGet, "/v1/advisors", a.listadvisorHandler)
	router.HandlerFunc(http.MethodPost, "/v1/advisors", a.createAdvisoryHandler)
	router.HandlerFunc(http.MethodGet, "/v1/advisors/:id", a.displayadvisoryHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/advisors/:id", a.updateadvisorHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/advisors/:id", a.deleteadvisorHandler)

	//Courses
	router.HandlerFunc(http.MethodGet, "/v1/courses", a.listcourseHandler)
	router.HandlerFunc(http.MethodPost, "/v1/courses", a.createcourseHandler)
	router.HandlerFunc(http.MethodGet, "/v1/courses/:id", a.displaycourseHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/courses/:id", a.updatecourseHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/courses/:id", a.deletecourseHandler)

	//Grades
	router.HandlerFunc(http.MethodGet, "/v1/grades", a.listGradeHandler)
	router.HandlerFunc(http.MethodPost, "/v1/grades", a.createGradeHandler)
	router.HandlerFunc(http.MethodGet, "/v1/grades/:id", a.displayGradeHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/grades/:id", a.updateGradeHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/grades/:id", a.deleteGradeHandler)

	//sections
	router.HandlerFunc(http.MethodGet, "/v1/sections", a.listSectionHandler)
	router.HandlerFunc(http.MethodPost, "/v1/sections", a.createSectionHandler)
	router.HandlerFunc(http.MethodGet, "/v1/sections/:id", a.displaySectionHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/sections/:id", a.updateSectionHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/sections/:id", a.deleteSectionHandler)

	//metrics
	router.Handler(http.MethodGet, "/v1/observability/database/metrics", expvar.Handler())

	return a.metrics(a.recoverPanic(a.enableCORS(a.rateLimit(LoggingMiddleware(GzipMiddleware(datasize, router))))))

}
