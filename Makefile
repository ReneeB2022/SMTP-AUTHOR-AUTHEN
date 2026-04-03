include .envrc

## run: run the cmd/api application
.PHONY: run
run:
	@echo  'Running application…'
	@go run ./cmd/api -port=4000 -env=production -limiter-burst=5 -limiter-rps=2 \
	-limiter-enabled=true -db-dsn=${university_DB_DSN} -cors-trusted-origins="http://localhost:9000 http://localhost:9001"
	
.PHONY: db/psql
db/psql:
	@psql ${university_DB_DSN}

##metrics: http://localhost:4000/v1/observability/database/metrics

##hey metrics 
##Get method:
.PHONY: hey-student
hey-student:
	hey http://localhost:4000/v1/students

##http://localhost:9001
.PHONY: cors
cors:
	@echo  'Running application…'
	@go run ./cmd/examples/cors/basic 
	
##http://localhost:9000
.PHONY: preflight
preflight:
	@echo  'Running application…'
	@go run ./cmd/examples/cors/preflight

##gzip commands
.PHONY: gzip
gzip:
	@curl -i --compressed http://localhost:4000/v1/students

.PHONY: original
original: 
	@echo 'Storing uncompressed data'
	@curl http://localhost:4000/v1/students --output students.json

.PHONY: compress
compress:
	@echo 'Compressing request'
	@curl -i -H "Accept-Encoding: gzip" http://localhost:4000/v1/students --output response.gz

.PHONY: compare
compare:
	@echo 'Comparing Compressed and Uncompressed Request'
	@ls -lh response.gz students.json

##user commands
.PHONY: create-user
create-user:
	BODY='{"username":"Marcus Fuller", "email":"MF@ub.edu.bz", "password": "Roaringcreek1"}';\
	curl -i -d "$$BODY" localhost:4000/v1/users

.PHONY: activate
activate:
	curl -X PUT -d '{"token": ""}' localhost:4000/v1/users/activated

.PHONY: authenticate
authenticate:	
	BODY='{"email": "MF@ub.edu.bz", "password": "Roaringcreek1"}';\
	curl -i -d "$$BODY" localhost:4000/v1/tokens/authentication

.PHONY: aut-All-students
aut-All-students:
	curl -i -H "Authorization: Bearer " localhost:4000/v1/students

##will fail because they do not have permission to write to the database
.PHONY: aut-create-student 
aut-create-student:

.PHONY: Freya-aut
Freya-aut:
	BODY='{"email": "Freya@example.com", "password": "Mistletoe"}';\
	curl -i -d "$$BODY" localhost:4000/v1/tokens/authentication

##will succeed because Freya Has permission to write to the database
.PHONY: Freya-write
Freya-write:
	BODY='{"Fname":"Pedro","Lname":"Pascal", "gender": "Male", "age": 45, "district_id": 3, "program": 12, "gpa": 3.25}';\
	curl -i -d "$$BODY" -H "Authorization: Bearer E6LOVG3MKX74CHDONRG5Q2OZEQ" localhost:4000/v1/students

.PHONY: Freya-Read
Freya-Read: 
	curl -i -H "Authorization: Bearer E6LOVG3MKX74CHDONRG5Q2OZEQ" localhost:4000/v1/students

## Student endpoint commands
.PHONY: create-student
create-student:
	@BODY='{"Fname":"Patrick","Lname":"Starr", "gender": "Male", "age": 17, "district_id": 1, "program": 10, "gpa": 2.40}'; \
	curl -i -d "$$BODY" localhost:4000/v1/students

.PHONY: All-students
All-students:
	curl -i http://localhost:4000/v1/students

.PHONY: one-student
one-student:
	@curl -i http://localhost:4000/v1/students/41

.PHONY: edit-student
edit-student:
	@curl -X PATCH -d '{"Lname": "Gomez", "age": 32, "gpa": 3.70}' localhost:4000/v1/students/41

.PHONY: delete-student
delete-student:
	@curl -X DELETE localhost:4000/v1/students/41

.PHONY: limit
limit: 
	for i in {1..8}; do curl -i localhost:4000/v1/students/40; done

.PHONY: pagination
pagination:
	curl -i "localhost:4000/v1/students?page=1&page_size=5"

.PHONY: asc
asc:
	@echo 'Ascending by student ID'
	@curl -i "localhost:4000/v1/students?sort=student_id";\

	@echo 'Ascending by First name'
	@curl -i "localhost:4000/v1/students?sort=Fname"

.PHONY: desc
desc:
	@echo 'Descending by Student ID'
	@curl -i "localhost:4000/v1/students?sort=-student_id";\

	@echo 'Descending by District'
	@2curl -i "localhost:4000/v1/students?sort=-District"

##Programs commands
.PHONY: create-program
create-program:
	@BODY='{"program":"Bachelors in Spanish","Procode":"BSPA", "Deptid": 2}'; \
	curl -i -d "$$BODY" localhost:4000/v1/programs

.PHONY: All-program
All-program:
	@curl -i http://localhost:4000/v1/programs

.PHONY: one-program
one-program:
	@curl -i http://localhost:4000/v1/programs/22

.PHONY: edit-program
edit-program:
	@curl -X PATCH -d '{"Deptid": 4}' localhost:4000/v1/programs/22

.PHONY: delete-program
delete-program:
	@curl -X DELETE localhost:4000/v1/programs/22

##Departments commands
.PHONY: create-dept
create-dept:
	@BODY='{"Deptname":"Faculty of Music and Composition","Deptcode":"FMC"}'; \
	curl -i -d "$$BODY" localhost:4000/v1/departments

.PHONY: All-departments
All-departments:
	@curl -i http://localhost:4000/v1/departments

.PHONY: one-department
one-department:
	@curl -i http://localhost:4000/v1/departments/4

.PHONY: edit-department
edit-department:
	@curl -X PATCH -d '{"Deptname": "Faculty of Social Sciences", "Deptcode": "FSS"}' localhost:4000/v1/departments/4

.PHONY: delete-department
delete-department:
	@curl -X DELETE localhost:4000/v1/departments/4
##Staff commands
.PHONY: create-staff
create-staff:
	@BODY='{"Fname":"Janice","Lname":"Mayweather", "role":2 , "department": 2}'; \
	curl -i -d "$$BODY" localhost:4000/v1/staff

.PHONY: All-staff
All-staff:
	@curl -i http://localhost:4000/v1/staff

.PHONY: one-staff
one-staff:
	@curl -i http://localhost:4000/v1/staff/1

.PHONY: edit-staff
edit-staff:
	@curl -X PATCH -d '{"role": 1}' localhost:4000/v1/staff/1

.PHONY: delete-staff
delete-staff:
	@curl -X DELETE localhost:4000/v1/staff/4

##Role Commands

.PHONY: create-role
create-role:
	@BODY='{"role":"Accountant"}'; \
	curl -i -d "$$BODY" localhost:4000/v1/roles

.PHONY: All-roles
All-roles:
	@curl -i http://localhost:4000/v1/roles

.PHONY: one-roles
one-roles:
	@curl -i http://localhost:4000/v1/roles/2

.PHONY: edit-roles
edit-roles:
	@curl -X PATCH -d '{"role": "Manager"}' localhost:4000/v1/roles/2

.PHONY: delete-roles
delete-roles:
	@curl -X DELETE localhost:4000/v1/roles/2

##District commands
.PHONY: create-district
create-district:
	@BODY='{"district":"San Pedro"}'; \
	curl -i -d "$$BODY" localhost:4000/v1/district

.PHONY: All-district
All-district:
	@curl -i http://localhost:4000/v1/district

.PHONY: one-district
one-district:
	@curl -i http://localhost:4000/v1/district/6

.PHONY: edit-district
edit-district:
	@curl -X PATCH -d '{"district": "Toledos"}' localhost:4000/v1/district/6

.PHONY: delete-district
delete-district:
	@curl -X DELETE localhost:4000/v1/district/6

##Courses
.PHONY: create-course
create-course:
	@BODY='{"Course":"College English II","Code":"ENGL1025", "credits":3 , "description": "This course develops literary interpretation, argumentation and research capabilities.", "prereq": "ENGL1014" , "fee": 145.20, "DeptId": 2}'; \
	curl -i -d "$$BODY" localhost:4000/v1/courses

.PHONY: All-course
All-course:
	@curl -i http://localhost:4000/v1/courses

.PHONY: one-course
one-course:
	@curl -i http://localhost:4000/v1/courses/6

.PHONY: edit-course
edit-course:
	@curl -X PATCH -d '{"available": 20, "fee": 191.20}' localhost:4000/v1/courses/6

.PHONY: delete-course
delete-course:
	@curl -X DELETE localhost:4000/v1/courses/6

##Grades
.PHONY: create-grade
create-grade:
	@BODY='{"Letter":"TBD","Value":0.01}'; \
	curl -i -d "$$BODY" localhost:4000/v1/grades

.PHONY: All-grade
All-grade:
	@curl -i http://localhost:4000/v1/grades

.PHONY: one-grade
one-grade:
	@curl -i http://localhost:4000/v1/grades/6

.PHONY: edit-grade
edit-grade:
	@curl -X PATCH -d '{"Value": 3.50}' localhost:4000/v1/courses/6

.PHONY: delete-grade
delete-grade:
	@curl -X DELETE localhost:4000/v1/grades/6

##Sections
.PHONY: create-section
create-section:
	@BODY='{"Lecturer":12,"Course_id":4, "available": 30, "Room": "Jabiru-U1", "Day": "MW", "starts": "2026-01-01T15:30:00Z", "finish": "2026-01-01T16:45:00Z", "semester": "2S25"}'; \
	curl -i -d "$$BODY" localhost:4000/v1/sections

.PHONY: All-section
All-section:
	@curl -i http://localhost:4000/v1/sections

.PHONY: one-section
one-section:
	@curl -i http://localhost:4000/v1/sections/6

.PHONY: edit-section
edit-section:
	@curl -X PATCH -d '{"Value": 3.50}' localhost:4000/v1/sections/6

.PHONY: delete-section
delete-section:
	@curl -X DELETE localhost:4000/v1/sections/6

##Advisor
.PHONY: create-advisory
create-advisory:
	@BODY='{"advisor":12,"student":19}'; \
	curl -i -d "$$BODY" localhost:4000/v1/advisors

.PHONY: one-advisory
one-advisory:
	@curl -i http://localhost:4000/v1/advisors/2

.PHONY: All-advisory
All-advisory:
	@curl -i http://localhost:4000/v1/advisors

.PHONY: edit-advisory
edit-advisory:
	@curl -X PATCH -d '{"advisor": 8}' localhost:4000/v1/advisors/2

.PHONY: delete-advisory
delete-advisory:
	@curl -X DELETE localhost:4000/v1/advisors/3

##enroll
.PHONY: create-enroll
create-enroll:
	@BODY='{"student":20,"section":6, "enrolled": "2026-01-19T15:30:00Z", "grade": 7}'; \
	curl -i -d "$$BODY" localhost:4000/v1/enrollments

.PHONY: one-enroll
one-enroll:
	@curl -i http://localhost:4000/v1/enrollments/2

.PHONY: All-enroll
All-enroll:
	@curl -i http://localhost:4000/v1/enrollments

.PHONY: edit-enroll
edit-enroll:
	@curl -X PATCH -d '{"grade": 1 }' localhost:4000/v1/enrollments/2

.PHONY: delete-enroll
delete-enroll:
	@curl -X DELETE localhost:4000/v1/enrollments/3


## migrations
##create new migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${university}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up:
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${university_DB_DSN} up

## db/migrations/down: apply all down database migrations
.PHONY: db/migrations/down
db/migrations/down:
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${university_DB_DSN} down
