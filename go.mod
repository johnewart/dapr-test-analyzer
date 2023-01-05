module test-analyzer

go 1.19

require (
	github.com/google/go-github/v48 v48.2.0
	github.com/gorilla/mux v1.8.0
	github.com/joho/godotenv v1.4.0
	golang.org/x/oauth2 v0.0.0-20180821212333-d2e6202438be
	gorm.io/driver/sqlite v1.4.3
	gorm.io/gorm v1.24.2
)

require (
	github.com/golang/protobuf v1.3.2 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-sqlite3 v1.14.15 // indirect
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5 // indirect
	golang.org/x/net v0.0.0-20210226172049-e18ecbb05110 // indirect
	google.golang.org/appengine v1.6.7 // indirect
)

replace github.com/google/go-github/v48 => ./lib/go-github
