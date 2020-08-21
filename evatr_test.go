package evatr

import (
	"testing"
)

func TestGermanVat(t *testing.T) {
	testCases := []struct {
		vat   string
		valid bool
	}{
		{`DE123456789`, true},
		{`DE 123456789`, true},
		{` DE 1234 56789 `, true},
		{`123456789`, false},
		{``, false},
		{`IR1234567WA`, false},
		{"LU 26375245", false},
	}

	for _, tc := range testCases {
		res := validateGermanVatNumber(removeWhiteSpace(tc.vat))
		if res != tc.valid {
			t.Errorf("validateGermanVatNumber = %s", tc.vat)
		}
	}
}

func TestForeignVat(t *testing.T) {
	testCases := []struct {
		vat   string
		valid bool
	}{
		{`DE123456789`, false},
		{`DE 123456789`, false},
		{` DE 1234 56789 `, false},
		{`123456789`, false},
		{``, false},
		{`IR1234567WA`, false},
		{"LU 26375245", true},
	}

	for _, tc := range testCases {
		res := validateForeignVatNumber(removeWhiteSpace(tc.vat))
		if res != tc.valid {
			t.Errorf("validateForeignVatNumber = %s", tc.vat)
		}
	}
}

func TestCheckSimpleError(t *testing.T) {
	input := &EvatrSimpleInput{
		OwnVatNumber:      "DE123456789",
		ValidateVatNumber: "LU 26375245",
	}
	resp, err := CheckSimple(input)
	if err != nil {
		t.Errorf("TestCheckSimple -  user: %s test: %s", input.OwnVatNumber, input.ValidateVatNumber)
	}

	if resp.ErrorCode != 206 {
		t.Errorf("TestCheckSimple did not error")
	}
}

func TestCheckSimpleValid(t *testing.T) {
	input := &EvatrSimpleInput{
		OwnVatNumber:      "DE115235681",
		ValidateVatNumber: "LU 26375245",
	}
	resp, err := CheckSimple(input)
	if err != nil {
		t.Errorf("TestCheckSimple -  user: %s test: %s", input.OwnVatNumber, input.ValidateVatNumber)
	}
	if !resp.IsValid {
		t.Errorf("TestCheckSimple should validate")
	}
}

func TestCheckQualifiedValid(t *testing.T) {
	input := &EvatrQualifiedInput{
		OwnVatNumber:      "DE 212 442 423",
		ValidateVatNumber: "LU 26375245",
		ValidateName:      "AMAZON EUROPE CORE S.A R.L.",
		ValidateCity:      "Luxembourg",
	}

	resp, err := CheckQualified(input)
	if err != nil {
		t.Errorf("CheckQualified -  user: %s test: %s", input.OwnVatNumber, input.ValidateVatNumber)
	}
	if resp.MatchedName != EvatrResponseMatched {
		t.Errorf("Name should match")
	}
	if resp.MatchedCity != EvatrResponseMatched {
		t.Errorf("City should match")
	}
}

func TestCheckQualifiedError(t *testing.T) {
	input := &EvatrQualifiedInput{
		OwnVatNumber:      "DE 212 442 423",
		ValidateVatNumber: "LU 26375245",
		ValidateName:      "Some Company",
		ValidateCity:      "Berlin",
	}

	resp, err := CheckQualified(input)
	if err != nil {
		t.Errorf("CheckQualified -  user: %s test: %s", input.OwnVatNumber, input.ValidateVatNumber)
	}
	if resp.MatchedName == EvatrResponseMatched {
		t.Errorf("Name should not match")
	}
}
