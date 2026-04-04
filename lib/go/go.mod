module github.com/nangashi/bmkr/lib/go

go 1.26.1

require (
	connectrpc.com/connect v1.19.1
	github.com/jackc/pgx/v5 v5.8.0
	github.com/labstack/echo/v4 v4.15.1
	github.com/nangashi/bmkr/gen/go v0.0.0-00010101000000-000000000000
	google.golang.org/protobuf v1.36.11
)

replace github.com/nangashi/bmkr/gen/go => ../../gen/go

require (
	github.com/labstack/gommon v0.4.2 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	golang.org/x/crypto v0.49.0 // indirect
	golang.org/x/net v0.52.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.35.0 // indirect
)
