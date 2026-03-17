module github.com/nangashi/bmkr/services/product-mgmt

go 1.26.1

replace github.com/nangashi/bmkr/gen/go => ../../gen/go

require github.com/jackc/pgx/v5 v5.8.0

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	golang.org/x/text v0.29.0 // indirect
)
