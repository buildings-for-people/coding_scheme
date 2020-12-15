package main

import (
	"testing"

	scheme_pkg "github.com/buildings-for-people/coding_scheme_object"
)

func TestLinksAreConsistent(t *testing.T) {

	var s scheme_pkg.Scheme

	err := s.ReadDomainFile("test_data/domain_1.md")
	if err != nil {
		t.Error(err.Error())
	}

	ln := "123 456"

	err = linksAreConsistent(ln, &s)

	if err != nil {
		t.Error(err.Error())
		t.FailNow()
	}

	ln = "123 456 [auto](layer=layer_1)"
	err = linksAreConsistent(ln, &s)
	if err != nil {
		t.Error(err.Error())
		t.FailNow()
	}

	ln = "123 456 ([auto](layer=layer_1))"
	err = linksAreConsistent(ln, &s)
	if err != nil {
		t.Error(err.Error())
		t.FailNow()
	}

	ln = "123 456 ([auto](layer=layer_1))  [auto](layer=layer_1)"
	err = linksAreConsistent(ln, &s)
	if err != nil {
		t.Error(err.Error())
		t.FailNow()
	}

	ln = "123 456 ([auto](layer=layer_1))  [auto](layer=layer_7)"
	err = linksAreConsistent(ln, &s)
	if err == nil {
		t.Error("Expecting non consistency")
		t.FailNow()
	}

	ln = "123 456 ([auto](layer=layer_7))  [auto](layer=layer_1)"
	err = linksAreConsistent(ln, &s)
	if err == nil {
		t.Error("Expecting non consistency")
		t.FailNow()
	}
}
