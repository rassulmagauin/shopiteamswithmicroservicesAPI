package data

import (
	"fmt"
	hclog "github.com/hashicorp/go-hclog"
	"testing"
)

func TestNewRates(t *testing.T) {
	er, err := NewRates(hclog.Default())
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("Rates %#v", er.rates)
}
