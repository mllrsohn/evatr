# eVatR: Bestätigung von ausländischen Umsatzsteuer-Identifikationsnummern

> Validate EU Vat Numbers with a valid German VAT Number

Uses the official API from the [Bundeszentralamt für Steuern](https://evatr.bff-online.de/eVatR/index_html).

It has a simple and qualified Check. This [Status Code](https://evatr.bff-online.de/eVatR/xmlrpc/codes) are always `200` even if other params don't match in the qualified check

## Install

```
go get -u github.com/mllrsohn/evatr
```

## Usage with Go

```go
import "github.com/mllrsohn/evatr"

input := &EvatrSimpleInput{
	OwnVatNumber:      "DEXXXXXXX",
	ValidateVatNumber: "LU26375245",
}
resp, err := CheckSimple(input)
if err != nil {
  // do sth with err
}

fmt.Println("isValid", resp.Valid)
	
```

```go
import "github.com/mllrsohn/evatr"

input := &EvatrQualifiedInput{
	OwnVatNumber:      "DEXXXXX",
    ValidateVatNumber: "LU26375245",
	ValidateName:      "AMAZON EUROPE CORE S.A R.L.",
	ValidateCity:      "Luxembourg",
}
resp, err := CheckQualified(input)
if err != nil {
  // do sth with err
}

if resp.MatchedName == EvatrResponseMatched {
    fmt.Println("Name and ID are a matching")
}
	
```
