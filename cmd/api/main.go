package main

import (
	"context"
	"database/sql"
	"expvar"
	"flag"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/ReneeB2022/test1/internal/mailer"

	"time"

	// the '_' means that we will not direct use the pq package
	"github.com/ReneeB2022/test1/internal/data"
	_ "github.com/lib/pq"
)

const appVersion = "1.0.0"

type serverConfig struct {
	port        int
	environment string
	db          struct {
		dsn string
	}
	limiter struct {
		rps     float64 // requests per second
		burst   int     // initial requests possible
		enabled bool    // enable or disable rate limiter
	}
	cors struct {
		trustedOrigins []string
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
}

type applicationDependencies struct {
	config          serverConfig
	logger          *slog.Logger
	departmentModel data.DepartmentModel
	studentModel    data.StudentModel
	rolemodel       data.RoleModel
	staffModel      data.StaffModel
	districtModel   data.DistrictModel
	programModel    data.ProgramModel
	enrollmentModel data.EnrollmentModel
	advisoryModel   data.AdvisoryModel
	courseModel     data.CourseModel
	grademodel      data.GradeModel
	sectionModel    data.SectionModel
	userModel       data.UserModel
	mailer          mailer.Mailer
	wg              sync.WaitGroup // need this later for background jobs
	tokenModel      data.TokenModel
	permissionModel data.PermissionModel
}

func main() {
	var settings serverConfig

	flag.IntVar(&settings.port, "port", 4000, "Server port")
	flag.StringVar(&settings.environment, "env", "development", "Environment(development|staging|production)")
	// read in the dsn
	flag.StringVar(&settings.db.dsn, "db-dsn", "postgres://university:password@localhost/university", "PostgreSQL DSN")

	flag.Float64Var(&settings.limiter.rps, "limiter-rps", 2,
		"Rate Limiter maximum requests per second")

	flag.IntVar(&settings.limiter.burst, "limiter-burst", 5,
		"Rate Limiter maximum burst")

	flag.BoolVar(&settings.limiter.enabled, "limiter-enabled", true,
		"Enable rate limiter")

	flag.StringVar(&settings.smtp.host,
		"smtp-host", "sandbox.smtp.mailtrap.io", "SMTP host")
	// We have port 25, 465, 587, 2525. If 25 doesn't work choose another
	flag.IntVar(&settings.smtp.port, "smtp-port", 25, "SMTP port")
	// Use your Username value provided by Mailtrap
	flag.StringVar(&settings.smtp.username, "smtp-username",
		"eda8776363d858", "SMTP username")

	flag.StringVar(&settings.smtp.password, "smtp-password",
		"3db35c0d79806a", "SMTP password")

	flag.StringVar(&settings.smtp.sender, "smtp-sender",
		"University <no-reply@university.reneebanner.net>",
		"SMTP sender")

	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)",
		func(val string) error {
			settings.cors.trustedOrigins = strings.Fields(val)
			return nil
		})

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// the call to openDB() sets up our connection pool
	db, err := openDB(settings)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	// release the database resources before exiting
	defer db.Close()

	logger.Info("database connection pool established")

	expvar.NewString("version").Set(appVersion)

	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	// the database connection pool metrics
	expvar.Publish("database", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	// the current Unix timestamp
	expvar.Publish("timestamp", expvar.Func(func() any {
		return time.Now().Unix()
	}))

	appInstance := &applicationDependencies{
		config:          settings,
		logger:          logger,
		departmentModel: data.DepartmentModel{DB: db},
		studentModel:    data.StudentModel{DB: db},
		rolemodel:       data.RoleModel{DB: db},
		staffModel:      data.StaffModel{DB: db},
		districtModel:   data.DistrictModel{DB: db},
		programModel:    data.ProgramModel{DB: db},
		enrollmentModel: data.EnrollmentModel{DB: db},
		advisoryModel:   data.AdvisoryModel{DB: db},
		courseModel:     data.CourseModel{DB: db},
		grademodel:      data.GradeModel{DB: db},
		sectionModel:    data.SectionModel{DB: db},
		userModel:       data.UserModel{DB: db},
		tokenModel:      data.TokenModel{DB: db},
		permissionModel: data.PermissionModel{DB: db},

		mailer: mailer.New(settings.smtp.host, settings.smtp.port,
			settings.smtp.username, settings.smtp.password, settings.smtp.sender),
	}
	err = appInstance.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func openDB(settings serverConfig) (*sql.DB, error) {
	// open a connection pool
	db, err := sql.Open("postgres", settings.db.dsn)
	if err != nil {
		return nil, err
	}

	// set a context to ensure DB operations don't take too long
	ctx, cancel := context.WithTimeout(context.Background(),
		5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	// return the connection pool (sql.DB)
	return db, nil

}
