package evatr

import (
	"encoding/xml"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var Timeout = 10 // seconds

var (
	ErrServiceNotReachable = errors.New("the evatr Service is offline")
	ErrInvalidGermanVat    = errors.New("input: the German VAT Number is invalid")
	ErrInvalidForeignVat   = errors.New("input: the Foreign VAT Number is invalid")
)

type EvatrSimpleResponse struct {
	ValidFrom string
}

const evatrURL = "https://evatr.bff-online.de/evatrRPC?"

type XMLResponse struct {
	Param []struct {
		Value struct {
			Array struct {
				Data struct {
					Value []struct {
						String string `xml:"string"`
					} `xml:"value"`
				} `xml:"data"`
			} `xml:"array"`
		} `xml:"value"`
	} `xml:"param"`
}

type EvatrResponseSimple struct {
	OwnVatNumber       string
	ValidatedVatNumber string
	ErrorCode          int
	IsValid            bool
}

type EvatrResponseMatch string

const (
	EvatrResponseMatched    EvatrResponseMatch = "A: Matched"
	EvatrResponseNotMatched EvatrResponseMatch = "B: No match"
	EvatrResponseNotQueried EvatrResponseMatch = "C: Not Queried"
	EvatrResponseUnknown    EvatrResponseMatch = "D: Not in Database"
	EvatrResponseInvalid    EvatrResponseMatch = "E: Invalid"
)

type EvatrResponseQualified struct {
	OwnVatNumber       string
	ValidatedVatNumber string
	ErrorCode          int
	MatchedName        EvatrResponseMatch
	MatchedCity        EvatrResponseMatch
	MatchedPostCode    EvatrResponseMatch
	MatchedStreet      EvatrResponseMatch
	IsValid            bool
}

type EvatrSimpleInput struct {
	OwnVatNumber      string
	ValidateVatNumber string
}

type EvatrQualifiedInput struct {
	OwnVatNumber      string
	ValidateVatNumber string
	ValidateName      string
	ValidateCity      string
	ValidatePostCode  string
	ValidateStreet    string
	ValidatePrint     bool
}

func CheckSimple(input *EvatrSimpleInput) (*EvatrResponseSimple, error) {
	// Check if the nubers look valid
	UstId_1 := removeWhiteSpace(input.OwnVatNumber)
	valid := validateGermanVatNumber(UstId_1)
	if !valid {
		return nil, ErrInvalidGermanVat
	}
	UstId_2 := removeWhiteSpace(input.ValidateVatNumber)
	valid = validateForeignVatNumber(UstId_2)
	if !valid {
		return nil, ErrInvalidForeignVat
	}

	payload := url.Values{}
	payload.Add("UstId_1", UstId_1)
	payload.Add("UstId_2", UstId_2)

	client := http.Client{
		Timeout: time.Duration(time.Duration(Timeout) * time.Second),
	}

	res, err := client.Get(evatrURL + payload.Encode())
	if err != nil {
		return nil, ErrServiceNotReachable
	}
	defer res.Body.Close()
	xmlRes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	respData := &XMLResponse{}
	err = xml.Unmarshal(xmlRes, respData) // Parse the XML into the structure
	if err != nil {
		return nil, err
	}

	// https://evatr.bff-online.de/eVatR/xmlrpc/codes
	errorCode := getValueFromReponse("ErrorCode", respData)
	errorCodeNumber := 0
	if errorCode != "" {
		errorCodeNumber, _ = strconv.Atoi(errorCode)
	}
	reponse := &EvatrResponseSimple{
		OwnVatNumber:       getValueFromReponse("UstId_1", respData),
		ValidatedVatNumber: getValueFromReponse("UstId_2", respData),
		ErrorCode:          errorCodeNumber,
		IsValid:            errorCodeNumber == 200,
	}

	return reponse, nil
}

func CheckQualified(input *EvatrQualifiedInput) (*EvatrResponseQualified, error) {
	// Check if the nubers look valid
	UstId_1 := removeWhiteSpace(input.OwnVatNumber)
	valid := validateGermanVatNumber(UstId_1)
	if !valid {
		return nil, ErrInvalidGermanVat
	}
	UstId_2 := removeWhiteSpace(input.ValidateVatNumber)
	valid = validateForeignVatNumber(UstId_2)
	if !valid {
		return nil, ErrInvalidForeignVat
	}

	payload := url.Values{}
	payload.Add("UstId_1", UstId_1)
	payload.Add("UstId_2", UstId_2)
	payload.Add("Firmenname", input.ValidateName)
	payload.Add("Ort", input.ValidateCity)
	payload.Add("PLZ", input.ValidatePostCode)
	print := "nein"
	if input.ValidatePrint {
		print = "ja"
	}
	payload.Add("Druck", print)

	client := http.Client{
		Timeout: time.Duration(time.Duration(Timeout) * time.Second),
	}

	res, err := client.Get(evatrURL + payload.Encode())
	if err != nil {
		return nil, ErrServiceNotReachable
	}
	defer res.Body.Close()
	xmlRes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	respData := &XMLResponse{}
	err = xml.Unmarshal(xmlRes, respData) // Parse the XML into the structure
	if err != nil {
		return nil, err
	}

	// https://evatr.bff-online.de/eVatR/xmlrpc/codes
	errorCode := getValueFromReponse("ErrorCode", respData)
	errorCodeNumber := 0
	if errorCode != "" {
		errorCodeNumber, _ = strconv.Atoi(errorCode)
	}
	reponse := &EvatrResponseQualified{
		OwnVatNumber:       getValueFromReponse("UstId_1", respData),
		ValidatedVatNumber: getValueFromReponse("UstId_2", respData),
		MatchedName:        getMatchStatus("Erg_Name", respData),
		MatchedCity:        getMatchStatus("Erg_Ort", respData),
		MatchedPostCode:    getMatchStatus("Erg_PLZ", respData),
		MatchedStreet:      getMatchStatus("Erg_Str", respData),
		ErrorCode:          errorCodeNumber,
	}

	return reponse, nil
}

func getMatchStatus(s string, respData *XMLResponse) EvatrResponseMatch {
	status := getValueFromReponse(s, respData)
	switch status {
	case "A":
		return EvatrResponseMatched
	case "B":
		return EvatrResponseNotMatched
	case "C":
		return EvatrResponseNotQueried
	case "D":
		return EvatrResponseUnknown
	case "":
	default:
		return EvatrResponseInvalid
	}

	return EvatrResponseInvalid
}

func getValueFromReponse(s string, respData *XMLResponse) string {
	for _, el := range respData.Param {
		for _, val := range el.Value.Array.Data.Value {
			if val.String == s {
				return el.Value.Array.Data.Value[1].String
			}
		}
	}
	return ""
}

func removeWhiteSpace(s string) string {
	rr := make([]rune, 0, len(s))
	for _, r := range s {
		if !unicode.IsSpace(r) {
			rr = append(rr, r)
		}
	}
	return string(rr)
}

func validateGermanVatNumber(n string) bool {
	if len(n) < 3 {
		return false
	}

	n = strings.ToUpper(n)
	if n[0:2] != "DE" {
		return false
	}
	matched, _ := regexp.MatchString("[0-9]{9}", n[2:])
	return matched
}

// https://github.com/dannyvankooten/vat/blob/master/numbers.go
func validateForeignVatNumber(n string) bool {
	patterns := map[string]string{
		"AT": `U[A-Z0-9]{8}`,
		"BE": `(0[0-9]{9}|[0-9]{10})`,
		"BG": `[0-9]{9,10}`,
		"CH": `(?:E(?:-| )[0-9]{3}(?:\.| )[0-9]{3}(?:\.| )[0-9]{3}( MWST)?|E[0-9]{9}(?:MWST)?)`,
		"CY": `[0-9]{8}[A-Z]`,
		"CZ": `[0-9]{8,10}`,
		"DK": `[0-9]{8}`,
		"EE": `[0-9]{9}`,
		"EL": `[0-9]{9}`,
		"ES": `[A-Z][0-9]{7}[A-Z]|[0-9]{8}[A-Z]|[A-Z][0-9]{8}`,
		"FI": `[0-9]{8}`,
		"FR": `([A-Z]{2}|[0-9]{2})[0-9]{9}`,
		"GB": `[0-9]{9}|[0-9]{12}|(GD|HA)[0-9]{3}`,
		"HR": `[0-9]{11}`,
		"HU": `[0-9]{8}`,
		"IE": `[A-Z0-9]{7}[A-Z]|[A-Z0-9]{7}[A-W][A-I]`,
		"IT": `[0-9]{11}`,
		"LT": `([0-9]{9}|[0-9]{12})`,
		"LU": `[0-9]{8}`,
		"LV": `[0-9]{11}`,
		"MT": `[0-9]{8}`,
		"NL": `[0-9]{9}B[0-9]{2}`,
		"PL": `[0-9]{10}`,
		"PT": `[0-9]{9}`,
		"RO": `[0-9]{2,10}`,
		"SE": `[0-9]{12}`,
		"SI": `[0-9]{8}`,
		"SK": `[0-9]{10}`,
	}

	if len(n) < 3 {
		return false
	}

	n = strings.ToUpper(n)
	pattern, ok := patterns[n[0:2]]
	if !ok {
		return false
	}
	matched, _ := regexp.MatchString(pattern, n[2:])
	return matched
}
