package api

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"sync"
	"time"
)

// Settings of the GopherNet.
type Settings struct {
	HTTPAddress        string
	UpdateStatusTicker time.Duration
	BurrowExpiration   time.Duration
	ReportStatusTicker time.Duration
	BurrowDigRate      float64
	Verbose            bool
}

// NewAPI creates a new API.
func NewAPI(s Settings, log *slog.Logger, burrows []Burrow) *API {
	return &API{
		settings: s,
		log:      log,
		burrows:  burrows,
	}
}

// API for the GopherNet.
type API struct {
	m        sync.Mutex
	log      *slog.Logger
	settings Settings
	burrows  []Burrow
}

// APIError response.
type APIError struct {
	HTTPCode int    `json:"http_code"`
	Message  string `json:"message"`
}

func (e APIError) Error() string {
	return e.Message
}

// Load API status.
func (a *API) Load(burrows []Burrow) {
	a.m.Lock()
	defer a.m.Unlock()
	a.burrows = burrows
}

// UpdateStatus updates the burrows status.
func (a *API) UpdateStatus(ctx context.Context) {
	if a.settings.UpdateStatusTicker == 0 {
		panic("update status ticker is not set")
	}

	ticker := time.NewTicker(a.settings.UpdateStatusTicker)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.updateStatus()
		}
	}
}

// ReportStatus writes the status of the burrows to a stream.
func (a *API) ReportStatus(ctx context.Context, w io.Writer) {
	a.reportStatus(w)
	ticker := time.NewTicker(a.settings.ReportStatusTicker)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.reportStatus(w)
		}
	}
}

func (a *API) reportStatus(w io.Writer) {
	if _, err := fmt.Fprintln(w, a.Report()); err != nil {
		a.log.Error("failed to write report file: %w", err)
	}
}

// updateStatus should be called as part of a background task routine, every minute.
func (a *API) updateStatus() {
	a.m.Lock()
	defer a.m.Unlock()

	for i, b := range a.burrows {
		a.burrows[i].Age++
		if b.Occupied {
			// Truncate the number to two decimal places for the sake of better readability.
			a.burrows[i].Depth = truncateNumber(a.burrows[i].Depth + a.settings.BurrowDigRate)
		}
	}

	// Remove burrows that are too old (have collapsed).
	a.burrows = slices.DeleteFunc(a.burrows, func(b Burrow) bool {
		return time.Duration(b.Age)*time.Minute > a.settings.BurrowExpiration
	})

	a.log.LogAttrs(context.Background(), slog.LevelDebug, "Burrows status updated", slog.Int("burrows", len(a.burrows)))
}

// truncateNumber "truncates" a float number to two decimal places.
// This is good enough for the purpose of this exercise.
// For real applications, consider using math/big to work with arbitrary-precision arithmetic.
func truncateNumber(n float64) float64 {
	f, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", n), 64)
	return f
}

// RentBurrow rents a burrow.
func (a *API) RentBurrow(name string, rented *Burrow) error {
	a.m.Lock()
	defer a.m.Unlock()

	for i := range a.burrows {
		if a.burrows[i].Name != name {
			continue
		}

		// Check if the burrow is available.
		if a.burrows[i].Occupied {
			return APIError{
				HTTPCode: http.StatusConflict,
				Message:  "burrow is already occupied",
			}
		}

		a.burrows[i].Occupied = true
		*rented = a.burrows[i]
		return nil
	}

	return APIError{
		HTTPCode: http.StatusNotFound,
		Message:  "burrow not found",
	}
}

// Status returns a list of current burrows.
func (a *API) Status(burrows *[]Burrow) {
	a.m.Lock()
	defer a.m.Unlock()
	*burrows = make([]Burrow, len(a.burrows))
	copy(*burrows, a.burrows)
}

// Report of the GopherNet.
func (a *API) Report() Report {
	report := Report{
		Time: time.Now(),
	}

	a.m.Lock()
	defer a.m.Unlock()

	var minVolume, maxVolume float64
	for _, b := range a.burrows {
		report.TotalDepth += b.Depth
		if !b.Occupied {
			report.Available++
		}

		volume := b.Volume()

		if report.Smallest == "" || volume < minVolume {
			report.Smallest = b.Name
			minVolume = volume
		}
		if report.Largest == "" || volume > maxVolume {
			report.Largest = b.Name
			maxVolume = volume
		}
	}

	return report
}
