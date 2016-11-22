package pacemaker_test

import (
	"fmt"
	"testing"
	"log"
	"strings"
	"bytes"
	"encoding/json"
	"github.com/krig/go-pacemaker"
)


func TestDecode(t *testing.T) {
	var err error

	cib := pacemaker.NewCibFile("testdata/simple.xml")
	defer cib.Delete()
	err = cib.SignOn(pacemaker.Query)
	if err != nil {
		t.Error(err)
	}
	defer cib.SignOff()

	err = cib.Decode()
	if err != nil {
		t.Error(err)
	}
}


func TestVersion(t *testing.T) {
	cib := pacemaker.NewCibFile("testdata/simple.xml")
	defer cib.Delete()
	err := cib.SignOn(pacemaker.Query)
	if err != nil {
		t.Error(err)
	}
	defer cib.SignOff()

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


func TestStatusJson(t *testing.T) {
	cib := pacemaker.NewCibFile("testdata/exit-reason.xml")
	defer cib.Delete()
	err := cib.SignOn(pacemaker.Query)
	if err != nil {
		t.Fatal(err)
	}
	defer cib.SignOff()

	err = cib.Decode()
	if err != nil {
		t.Fatal(err)
	}

	data, err := cib.Status.ToJson()
	if err != nil {
		t.Fatal(err)
	}

    var prettyJSON bytes.Buffer
    err = json.Indent(&prettyJSON, data, "", "  ")
    if err != nil {
		t.Fatal(err)
    }

	jsonstr := prettyJSON.String()

	if !strings.Contains(jsonstr, "\"id\": \"gctvanas-lvm\"") {
		t.Fatal("Expected gctvanas-lvm, got ", jsonstr)
	}
}

func TestCibJson(t *testing.T) {
	cib := pacemaker.NewCibFile("testdata/exit-reason.xml")
	defer cib.Delete()
	err := cib.SignOn(pacemaker.Query)
	if err != nil {
		t.Fatal(err)
	}
	defer cib.SignOff()

	err = cib.Decode()
	if err != nil {
		t.Fatal(err)
	}

	data, err := cib.ToJson()
	if err != nil {
		t.Fatal(err)
	}

    var prettyJSON bytes.Buffer
    err = json.Indent(&prettyJSON, data, "", "  ")
    if err != nil {
		t.Fatal(err)
    }

	jsonstr := prettyJSON.String()
	log.Printf("%s", jsonstr)

	if !strings.Contains(jsonstr, "gctvanas-fs2o-meta_attributes-clone-node-max") {
		t.Fatal("Expected configuration, got ", jsonstr)
	}
}


func ExampleQuery() {
	cib := pacemaker.NewCibFile("testdata/simple.xml")
	defer cib.Delete()
	err := cib.SignOn(pacemaker.Query)
	if err != nil {
		log.Fatal(err)
	}
	defer cib.SignOff()

	xml, err := cib.Query()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", xml[0:4])
	// Output: <cib
}

func ExampleQueryXPath() {
	cib := pacemaker.NewCibFile("testdata/simple.xml")
	defer cib.Delete()
	err := cib.SignOn(pacemaker.Query)
	if err != nil {
		log.Fatal(err)
	}
	defer cib.SignOff()

	xml, err := cib.QueryXPath("//nodes/node[@id=\"xxx\"]")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", xml)
	// Output: <node id="xxx" uname="c001n01" type="normal"/>
}

func ExampleDecode() {
	var err error

	cib := pacemaker.NewCibFile("testdata/simple.xml")
	defer cib.Delete()
	err = cib.SignOn(pacemaker.Query)
	if err != nil {
		log.Fatal(err)
	}
	defer cib.SignOff()

	err = cib.Decode()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%v", *cib.ValidateWith)
	// Output: pacemaker-1.2
}

func findOps(cib *pacemaker.Cib, nodename string, rscname string) []pacemaker.ResourceStateOp {
	for _, node := range cib.Status.NodeState {
		if node.Uname == nodename {
			for _, rsc := range node.Resources {
				if rsc.Id == rscname {
					return rsc.Ops
				}
			}
		}
	}
	return nil
}


func ExampleDecodeStatus() {
	var err error

	cib := pacemaker.NewCibFile("testdata/exit-reason.xml")
	defer cib.Delete()
	err = cib.SignOn(pacemaker.Query)
	if err != nil {
		log.Fatal(err)
	}
	defer cib.SignOff()

	err = cib.Decode()
	if err != nil {
		log.Fatal(err)
	}

	ops := findOps(cib, "node1", "gctvanas-lvm")
	fmt.Printf(ops[0].ExitReason)
	// Output: LVM: targetfs did not activate correctly
}
