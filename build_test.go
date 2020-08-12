package main

import "testing"

func TestLinksAreConsistent(t *testing.T) {

	ln := "123 456"
	domains := []string{"domain 1", "domain 2", "domain 3"}
	layers := []string{"layer 1", "layer 2"}
	codes := []string{"code 1", "code 2", "code 3", "code 4"}

	areConsistent, err := linksAreConsistent(ln, &domains, &layers, &codes)
	if !areConsistent {
		t.Error()
		t.FailNow()
	}
	if err != nil {
		t.Error(err.Error())
		t.FailNow()
	}

	ln = "123 456 [auto](layer=layer_1)"
	areConsistent, err = linksAreConsistent(ln, &domains, &layers, &codes)
	if err != nil {
		t.Error(err.Error())
		t.FailNow()
	}
	if !areConsistent {

		t.Error()
		t.FailNow()
	}

	ln = "123 456 ([auto](layer=layer_1))"
	areConsistent, err = linksAreConsistent(ln, &domains, &layers, &codes)
	if !areConsistent {
		t.Error()
		t.FailNow()
	}
	if err != nil {
		t.Error(err.Error())
		t.FailNow()
	}

	ln = "123 456 ([auto](layer=layer_1))  [auto](layer=layer_1)"
	areConsistent, err = linksAreConsistent(ln, &domains, &layers, &codes)
	if !areConsistent {
		t.Error()
		t.FailNow()
	}
	if err != nil {
		t.Error(err.Error())
		t.FailNow()
	}

	ln = "123 456 ([auto](layer=layer_1))  [auto](layer=layer_7)"
	areConsistent, err = linksAreConsistent(ln, &domains, &layers, &codes)
	if areConsistent {
		t.Error()
		t.FailNow()
	}
	if err == nil {
		t.Error("Expecting non consistency")
		t.FailNow()
	}

	ln = "123 456 ([auto](layer=layer_7))  [auto](layer=layer_1)"
	areConsistent, err = linksAreConsistent(ln, &domains, &layers, &codes)
	if areConsistent {
		t.Error()
		t.FailNow()
	}
	if err == nil {
		t.Error("Expecting non consistency")
		t.FailNow()
	}
}
