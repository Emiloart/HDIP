module github.com/Emiloart/HDIP/services/verifier-api

go 1.26

require (
	github.com/Emiloart/HDIP/packages/go/foundation v0.0.0
	github.com/Emiloart/HDIP/services/internal/phase1sql v0.0.0
	github.com/Emiloart/HDIP/services/internal/phase1sqltest v0.0.0
)

replace github.com/Emiloart/HDIP/packages/go/foundation => ../../packages/go/foundation

replace github.com/Emiloart/HDIP/services/internal/phase1sql => ../internal/phase1sql

replace github.com/Emiloart/HDIP/services/internal/phase1sqltest => ../internal/phase1sqltest
