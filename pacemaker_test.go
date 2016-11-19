package pacemaker_test

import (
	"os"
	"fmt"
	"flag"
	"testing"
	"github.com/krig/go-pacemaker"
)


func TestMain(m *testing.M) {
	flag.Parse()
	os.Setenv("CIB_file", "testdata/simple.xml")
	os.Exit(m.Run())
}


func ExampleVersion() {
	cib := pacemaker.NewCib()
	defer cib.Delete()
	err := cib.SignOn(pacemaker.Query)
	if err != nil {
		panic(err.Error())
	}
	defer cib.SignOff()

	ver, err := cib.Version()
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Version = %s\n", ver.String())
	// Output: Version = 1:0:0
}


func ExampleQuery() {
	cib := pacemaker.NewCib()
	defer cib.Delete()
	err := cib.SignOn(pacemaker.Query)
	if err != nil {
		panic(err.Error())
	}
	defer cib.SignOff()

	xml, err := cib.Query()
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("%s\n", xml[0:4])
	// Output: <cib
}

func ExampleQueryXPath() {
	cib := pacemaker.NewCib()
	defer cib.Delete()
	err := cib.SignOn(pacemaker.Query)
	if err != nil {
		panic(err.Error())
	}
	defer cib.SignOff()

	xml, err := cib.QueryXPath("//nodes/node[@id=\"xxx\"]")
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("%s\n", xml)
	// Output: <node id="xxx" uname="c001n01" type="normal"/>
}
