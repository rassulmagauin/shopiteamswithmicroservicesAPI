package data

import "testing"

func TestCheckValidation(t *testing.T) {
	p := &Product{
		Name: "aidyn",
		SKU:  "dsa-ddfsc-asx",
	}

	err := p.Validate()

	if err != nil {
		t.Fatal(err)
	}
}
