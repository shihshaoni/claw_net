package constraints

import "time"

type ConstraintSet struct {
	MaxRounds   int
	MaxMessages int
	MaxBudgetUSDCents int64
	Deadline    time.Duration
}

func Default() ConstraintSet {
	return ConstraintSet{
		MaxRounds:   12,
		MaxMessages: 200,
		MaxBudgetUSDCents: 500, // $5.00 placeholder (v0)
		Deadline:    10 * time.Second,
	}
}
