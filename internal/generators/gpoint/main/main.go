package main

import (
	"fmt"
	"os"

	"github.com/consensys/gurvy/internal/generators/curve"
	"github.com/consensys/gurvy/internal/generators/gpoint"
)

func main() {

	// g1.go
	{
		d := gpoint.Data{
			Fpackage:     curve.C.Fpackage,
			PName:        curve.C.PointName + "1",
			CoordType:    curve.C.FpName + ".Element", // TODO refer to other constants
			GroupType:    curve.C.FrName,
			ThirdRootOne: curve.C.ThirdRootOne,
			Lambda:       curve.C.Lambda,
			Size1:        curve.C.Size1,
			Size2:        curve.C.Size2,
		}

		// assume working directory is internal/generators
		// TODO make this path more robust to changes in working directory
		if err := gpoint.Generate(d, "../../"); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-1)
		}
	}

	// g2.go
	{
		// G2: sextic twist => CoordType is a field extension of degree EmbeddingDegree / 6
		var coordType string
		switch curve.C.EmbeddingDegree {
		case 12:
			coordType = curve.C.Fp2Name
		case 6:
			coordType = curve.C.FpName + ".Element" // TODO refer to other constants
		default:
			panic("embedding degree " + string(curve.C.EmbeddingDegree) + " not implemented")
		}

		d := gpoint.Data{
			Fpackage:  curve.C.Fpackage,
			PName:     curve.C.PointName + "2",
			CoordType: coordType,
			GroupType: curve.C.FrName,
			// ThirdRootOne: curve.C.ThirdRootOne,
			// Lambda:       curve.C.Lambda,
			// Size1:        curve.C.Size1,
			// Size2:        curve.C.Size2,
		}

		// assume working directory is internal/generators
		// TODO make this path more robust to changes in working directory
		if err := gpoint.Generate(d, "../../"); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-1)
		}
	}
}
