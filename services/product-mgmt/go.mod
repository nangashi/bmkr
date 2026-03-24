module github.com/nangashi/bmkr/services/product-mgmt

go 1.26.1

replace github.com/nangashi/bmkr/gen/go => ../../gen/go

require (
	connectrpc.com/connect v1.19.1
	github.com/a-h/templ v0.3.1001
	github.com/jackc/pgx/v5 v5.8.0
	github.com/labstack/echo/v4 v4.15.1
	github.com/nangashi/bmkr/gen/go v0.0.0-00010101000000-000000000000
	golang.org/x/net v0.52.0
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/labstack/gommon v0.4.2 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	golang.org/x/crypto v0.49.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.35.0 // indirect
	golang.org/x/time v0.14.0 // indirect
)
