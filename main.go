package main

import (
	sdk "github.com/hashicorp/waypoint-plugin-sdk"
	"github.com/peytoncasper/waypoint-plugin-sbt/builder"
)

func main() {
	sdk.Main(sdk.WithComponents(
		&builder.Builder{},
	))
}
