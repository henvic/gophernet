package api

import (
	"fmt"
	"math"
	"time"
)

// Burrow dig by a gopher.
type Burrow struct {
	Name     string  `json:"name"`     // Name of the burrow.
	Depth    float64 `json:"depth"`    // Depth of the burrow, in meters.
	Width    float64 `json:"width"`    // Width of the burrow, in meters.
	Occupied bool    `json:"occupied"` // Whether the burrow is occupied.
	Age      int     `json:"age"`      // Age of the burrow, in minutes.
}

// Volume of the burrow.
func (b Burrow) Volume() float64 {
	radius := b.Width * 0.5
	return b.Depth * math.Pow(radius, 2) * math.Pi
}

// Report of the GopherNet.
type Report struct {
	Time       time.Time // Time of the report.
	TotalDepth float64   // Total depth of all burrows.
	Available  int       // Number of available burrows.
	Largest    string    // Largest burrow.
	Smallest   string    // Smallest burrow.
}

func (r Report) String() string {
	return fmt.Sprintf("%v TotalDepth: %.2f, Available: %d, Largest: %s, Smallest: %s", r.Time.Format(time.DateTime), r.TotalDepth, r.Available, r.Largest, r.Smallest)
}
