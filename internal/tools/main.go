// Package main exists to provide imports for tools used in development
package main

import (
	_ "github.com/mgechev/revive"
	_ "github.com/nbutton23/zxcvbn-go" // required by gosec
	_ "github.com/securego/gosec/v2"
	_ "github.com/securego/gosec/v2/autofix"      // required by gosec
	_ "github.com/securego/gosec/v2/report/sarif" // required by gosec
	_ "github.com/securego/gosec/v2/report/text"  // required by gosec
	_ "github.com/swaggo/swag/cmd/swag"           // required by gosec
)
