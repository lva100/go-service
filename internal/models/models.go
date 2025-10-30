package models

import "time"

type Otkrep struct {
	PID        string    `db:"PID"`
	ENP        string    `db:"ENP"`
	LpuCodeNew string    `db:"LpuCodeNew"`
	LpuNameNew string    `db:"LpuNameNew"`
	LpuStart   time.Time `db:"LpuStart"`
	LpuFinish  time.Time `db:"LpuFinish"`
	LpuCode    string    `db:"LpuCode"`
}
