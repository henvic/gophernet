package api

import "testing"

func TestRentBurrow(t *testing.T) {
	api := NewAPI(Settings{}, nil, []Burrow{
		{Name: "burrow1", Occupied: false},
		{Name: "burrow2", Occupied: true},
		{Name: "burrow3", Occupied: false},
	})

	var burrow1 Burrow
	if err := api.RentBurrow("burrow1", &burrow1); err != nil {
		t.Errorf("RentBurrow returned an error: %v", err)
	}

	if burrow1.Name != "burrow1" {
		t.Errorf("expecteed rented burrow to be 'burrow1', got %v instead", burrow1.Name)
	}

	var burrow2 Burrow
	if err := api.RentBurrow("burrow2", &burrow2); err == nil {
		t.Errorf("expected error for occupied burrow")
	}

	var burrow4 Burrow
	if err := api.RentBurrow("burrow4", &burrow4); err == nil {
		t.Error("RentBurrow did not return an error for a non-existent burrow")
	}
}
