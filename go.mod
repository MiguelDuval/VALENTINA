module github.com/MiguelDuval/valentina

go 1.25.0

require (
	github.com/google/uuid v1.6.0
	github.com/oklog/ulid/v2 v2.1.1
	modernc.org/sqlite v1.46.1
)

replace github.com/scrypster/huginn => ./upstream/huginn
