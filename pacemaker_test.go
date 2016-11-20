package pacemaker_test

import (
	"os"
	"fmt"
	"flag"
	"testing"
	"log"
	"github.com/krig/go-pacemaker"
)


func TestMain(m *testing.M) {
	flag.Parse()
	os.Setenv("CIB_file", "testdata/simple.xml")
	os.Exit(m.Run())
}

func TestDecode(t *testing.T) {
	var err error

	cib := pacemaker.NewCib()
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
	cib := pacemaker.NewCib()
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


func ExampleQuery() {
	cib := pacemaker.NewCib()
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
	cib := pacemaker.NewCib()
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

	cib := pacemaker.NewCib()
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

	fmt.Printf("%v", cib.Attributes["validate-with"])
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

	os.Setenv("CIB_file", "testdata/exit-reason.xml")
	cib := pacemaker.NewCib()
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
