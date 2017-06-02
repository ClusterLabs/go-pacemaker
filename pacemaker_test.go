// Copyright (C) 2017 Kristoffer Gronlund <kgronlund@suse.com>
// See LICENSE for license.
package pacemaker_test

import (
	"fmt"
	"testing"
	"log"
	"strings"
	"github.com/krig/go-pacemaker"
	"gopkg.in/xmlpath.v2"
)


func TestXmlpath(t *testing.T) {
	cib, err := pacemaker.OpenCib(pacemaker.FromFile("testdata/simple.xml"))
	if err != nil {
		t.Fatal(err)
	}
	defer cib.Close()

	doc, err := cib.Query()
	if err != nil {
		t.Fatal(err)
	}
	defer doc.Close()

	path := xmlpath.MustCompile("/cib/configuration/nodes/node[@id='xxx']/@type")
	root, err := xmlpath.Parse(strings.NewReader(doc.ToString()))
	if err != nil {
        t.Fatal(err)
	}
	value, ok := path.String(root)
	if !ok {
		t.Error("xpath query failed")
	}
	if value != "normal" {
		t.Error(fmt.Sprintf("Expected 'normal', got '%v'", value))
	}
}


func TestVersion(t *testing.T) {
	cib, err := pacemaker.OpenCib(pacemaker.FromFile("testdata/simple.xml"))
	if err != nil {
		t.Fatal(err)
	}
	defer cib.Close()

	ver, err := cib.Version()
	if err != nil {
		t.Error(err)
	}

	if ver.AdminEpoch != 1 {
		t.Error("Expected admin_epoch == 1, got ", ver.AdminEpoch)
	}
	if ver.Epoch != 0 {
		t.Error("Expected epoch == 0, got ", ver.Epoch)
	}
}


func ExampleQuery() {
	cib, err := pacemaker.OpenCib(pacemaker.FromFile("testdata/simple.xml"))
	if err != nil {
		log.Fatal(err)
	}
	defer cib.Close()

	doc, err := cib.Query()
	if err != nil {
		log.Fatal(err)
	}
	defer doc.Close()

	fmt.Printf("%s\n", doc.ToString()[0:4])
	// Output: <cib
}

func ExampleQueryXPath() {
	cib, err := pacemaker.OpenCib(pacemaker.FromFile("testdata/simple.xml"))
	if err != nil {
		log.Fatal(err)
	}
	defer cib.Close()

	doc, err := cib.QueryXPath("//nodes/node[@id=\"xxx\"]")
	if err != nil {
		log.Fatal(err)
	}
	defer doc.Close()

	fmt.Printf("%s\n", doc.ToString())
	// Output: <node id="xxx" uname="c001n01" type="normal"/>
}
