package models

import "time"

type UslReport struct {
	Start    time.Time
	Code_MO  string
	OrgName  string
	Code_Usl string
	MC       int16
	MF       int16
	Usl_vol  float64
	Usl_fin  float64
}
