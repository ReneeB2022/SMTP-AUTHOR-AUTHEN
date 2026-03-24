CREATE TABLE IF NOT EXISTS Departments(
    department_id SERIAL PRIMARY KEY,
    department_name text NOT NULL,
    department_code text NOT NULL,
    created_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS Programs(
    program_id SERIAL PRIMARY KEY,
    Program text NOT NULL,
    program_code text NOT NULL,
    department_id integer REFERENCES Departments(department_id),
    created_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS Districts(
    district_id SERIAL PRIMARY KEY,
    district_name text NOT NULL,
    created_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS Roles(
    role_id SERIAL PRIMARY KEY,
    roles text NOT NULL,
    created_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS Staff(
    staff_id SERIAL PRIMARY KEY,
    Fname text NOT NULL,
    Lname text NOT NULL,
    Role_id integer REFERENCES Roles(role_id),
    Depart_id integer REFERENCES Departments(department_id),
    created_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS Students(
    student_id SERIAL PRIMARY KEY,
    Fname text NOT NULL,
    Lname text NOT NULL,
    Gender text NOT NULL,
    age integer NOT NULL,
    District integer REFERENCES Districts(district_id),
    program_id integer REFERENCES Programs(program_id),
    GPA decimal(3,2) NOT NULL,
    created_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS Advisors_Students(
    ADSTU_id SERIAL PRIMARY KEY,
    advisor_id integer REFERENCES Staff(staff_id),
    student_id integer REFERENCES Students(student_id),
    created_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS Courses(
    course_id SERIAL PRIMARY KEY,
    course_name text NOT NULL,
    course_code text NOT NULL,
    credits integer NOT NULL,
    Descriptions text,
    Prerequisites text,
    Fee integer NOT NULL,
    dept_id integer REFERENCES Departments(department_id),
    created_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS Sections(
  section_id serial PRIMARY KEY,
  staff_id integer REFERENCES Staff(staff_id),
  course_id integer REFERENCES Courses(course_id),
  availability integer NOT NULL,
  classroom text NOT NULL,
  classday text NOT NULL,
  starttime TIME WITHOUT TIME ZONE NOT NULL,
  endtime TIME WITHOUT TIME ZONE NOT NULL,
  semester text NOT NULL,
  created_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS Grades(
  grade_id serial PRIMARY KEY,
  Letter_grade text NOT NULL,
  grade_value NUMERIC(3,2) NOT NULL,
  created_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS Enrollments(
  enrollment_id serial PRIMARY KEY,
  student_id integer REFERENCES Students(student_id),
  section_id integer REFERENCES Sections(section_id),
  enrollment_date DATE NOT NULL,
  grade_id integer REFERENCES Grades(grade_id),
  created_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
);