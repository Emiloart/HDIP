module github.com/Emiloart/HDIP/services/internal/phase1sqltest

go 1.26

require (
	github.com/Emiloart/HDIP/services/internal/phase1sql v0.0.0
	modernc.org/sqlite v1.49.1
)

replace github.com/Emiloart/HDIP/services/internal/phase1sql => ../phase1sql
