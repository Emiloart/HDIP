module github.com/Emiloart/HDIP/services/trust-registry

go 1.26

require (
	github.com/Emiloart/HDIP/packages/go/foundation v0.0.0
	github.com/Emiloart/HDIP/services/internal/phase1runtime v0.0.0
	github.com/Emiloart/HDIP/services/internal/phase1sql v0.0.0
)

replace github.com/Emiloart/HDIP/packages/go/foundation => ../../packages/go/foundation

replace github.com/Emiloart/HDIP/services/internal/phase1runtime => ../internal/phase1runtime

replace github.com/Emiloart/HDIP/services/internal/phase1sql => ../internal/phase1sql
