package model

import (
	"fmt"
	"strings"
)

type ChemicalClass string

const (
	ClassAcid    ChemicalClass = "acid"
	ClassBase    ChemicalClass = "base"
	ClassSolvent ChemicalClass = "solvent"
	ClassWater   ChemicalClass = "water"
)

type Chemical struct {
	Name          string        `json:"name"`
	Class         ChemicalClass `json:"class"`
	Hazard        int           `json:"hazard"`
	TargetDensity float64       `json:"target_density"`
}

func (c Chemical) Validate() error {
	if strings.TrimSpace(c.Name) == "" {
		return fmt.Errorf("chemical name is required")
	}
	if c.Hazard < 1 || c.Hazard > 5 {
		return fmt.Errorf("hazard level must be between 1 and 5")
	}
	switch c.Class {
	case ClassAcid, ClassBase, ClassSolvent, ClassWater:
		return nil
	default:
		return fmt.Errorf("unsupported compatibility class %q", c.Class)
	}
}

func Compatible(left, right ChemicalClass) bool {
	if left == "" || right == "" {
		return false
	}
	if left == ClassWater || right == ClassWater {
		return true
	}
	return left == right
}

func Catalog() map[string]Chemical {
	return map[string]Chemical{
		"sulfuric":     {Name: "sulfuric", Class: ClassAcid, Hazard: 5, TargetDensity: 1.10},
		"hydrochloric": {Name: "hydrochloric", Class: ClassAcid, Hazard: 4, TargetDensity: 1.08},
		"hydrofluoric": {Name: "hydrofluoric", Class: ClassAcid, Hazard: 5, TargetDensity: 1.03},
		"nitric":       {Name: "nitric", Class: ClassAcid, Hazard: 5, TargetDensity: 1.12},
		"peroxide":     {Name: "peroxide", Class: ClassAcid, Hazard: 4, TargetDensity: 1.05},
		"ammonia":      {Name: "ammonia", Class: ClassBase, Hazard: 4, TargetDensity: 0.98},
		"solvent":      {Name: "solvent", Class: ClassSolvent, Hazard: 3, TargetDensity: 0.82},
	}
}
