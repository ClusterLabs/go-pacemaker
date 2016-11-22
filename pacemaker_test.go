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
	cib, err := pacemaker.OpenCib(pacemaker.FromFile("testdata/simple.xml"))
	defer cib.Close()
	if err != nil {
		t.Fatal(err)
	}

	err = cib.Decode()
	if err != nil {
		t.Fatal(err)
	}
}


func TestVersion(t *testing.T) {
	cib, err := pacemaker.OpenCib(pacemaker.FromFile("testdata/simple.xml"))
	defer cib.Close()
	if err != nil {
		t.Fatal(err)
	}

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
	cib, err := pacemaker.OpenCib(pacemaker.FromFile("testdata/exit-reason.xml"))
	defer cib.Close()
	if err != nil {
		t.Fatal(err)
	}

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
	cib, err := pacemaker.OpenCib(pacemaker.FromFile("testdata/exit-reason.xml"))
	defer cib.Close()
	if err != nil {
		t.Fatal(err)
	}

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
	cib, err := pacemaker.OpenCib(pacemaker.FromFile("testdata/simple.xml"))
	defer cib.Close()
	if err != nil {
		log.Fatal(err)
	}

	xml, err := cib.Query()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", xml[0:4])
	// Output: <cib
}

func ExampleQueryXPath() {
	cib, err := pacemaker.OpenCib(pacemaker.FromFile("testdata/simple.xml"))
	defer cib.Close()
	if err != nil {
		log.Fatal(err)
	}

	xml, err := cib.QueryXPath("//nodes/node[@id=\"xxx\"]")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", xml)
	// Output: <node id="xxx" uname="c001n01" type="normal"/>
}

func ExampleDecode() {
	cib, err := pacemaker.OpenCib(pacemaker.FromFile("testdata/simple.xml"))
	defer cib.Close()
	if err != nil {
		log.Fatal(err)
	}

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
	cib, err := pacemaker.OpenCib(pacemaker.FromFile("testdata/exit-reason.xml"))
	defer cib.Close()
	if err != nil {
		log.Fatal(err)
	}

	err = cib.Decode()
	if err != nil {
		log.Fatal(err)
	}

	//ops := findOps(cib, "node1", "gctvanas-lvm")
	//fmt.Printf(ops[0].ExitReason)
	// Output: LVM: targetfs did not activate correctly
}
