// The pacemaker package provides an API for reading the Pacemaker cluster configuration (CIB).
package pacemaker

import (
	"unsafe"
	"fmt"
	"encoding/xml"
)

/*
#cgo pkg-config: libxml-2.0 glib-2.0 libqb pacemaker pacemaker-cib
#include <crm/cib.h>
#include <crm/common/util.h>

int go_cib_signon(cib_t* cib, const char* name, enum cib_conn_type type) {
    return cib->cmds->signon(cib, name, type);
}

int go_cib_signoff(cib_t* cib) {
    return cib->cmds->signoff(cib);
}

int go_cib_query(cib_t * cib, const char *section, xmlNode ** output_data, int call_options) {
    return cib->cmds->query(cib, section, output_data, call_options);
}

*/
import "C"

// Error type returned by the functions in this package.
type CibError struct {
	msg string
}

func (e *CibError) Error() string {
	return e.msg
}

// Internal function used to create a CibError instance
// from a pacemaker return code.
func formatErrorRc(rc int) *CibError {
	errorname := C.pcmk_errorname(C.int(rc))
	strerror := C.pcmk_strerror(C.int(rc))
	if errorname == nil {
		errorname = C.CString("")
		defer C.free(unsafe.Pointer(errorname))
	}
	if strerror == nil {
		strerror = C.CString("")
		defer C.free(unsafe.Pointer(strerror))
	}
	return &CibError{fmt.Sprintf("%d: %s %s", rc, C.GoString(errorname), C.GoString(strerror))}
}

// When connecting to Pacemaker, we have
// to declare which type of connection to
// use. Since the API is read-only at the
// moment, it only really makes sense to
// pass Query to functions that take a
// CibConnection parameter.
type CibConnection int

const (
	Query CibConnection = C.cib_query
	Command CibConnection = C.cib_command
)

// Root entity representing the CIB. Can be
// populated with CIB data if the Decode
// method is used.
type Cib struct {
	cCib *C.cib_t

	ValidateWith string `xml:"validate-with,attr"`
	AdminEpoch int `xml:"admin_epoch,attr"`
	Epoch int `xml:"epoch,attr"`
	NumUpdates int `xml:"num_updates,attr"`
	CrmFeatureSet string `xml:"crm_feature_set,attr"`
	RemoteTlsPort int `xml:"remote-tls-port,attr"`
	RemoteClearPort int `xml:"remote-clear-port,attr"`
	HaveQuorum string `xml:"have-quorum,attr"`
	DcUuid string `xml:"dc-uuid,attr"`
	CibLastWritten string `xml:"cib-last-written,attr"`
	NoQuorumPanic string `xml:"no-quorum-panic,attr"`
	UpdateOrigin string `xml:"update-origin,attr"`
	UpdateClient string `xml:"update-client,attr"`
	UpdateUser string `xml:"update-user,attr"`
	ExecutionDate string `xml:"execution-date,attr"`
	Configuration Configuration `xml:"configuration"`
	Status Status `xml:"status"`
}

// Represents a configuration name-value pair.
type NVPair struct {
	Id *string `xml:"id,attr"`
	IdRef *string `xml:"id-ref,attr"`
	Name *string `xml:"name,attr"`
	Value *string `xml:"value,attr"`
}

type DateSpec struct {
	Id string `xml:"id,attr"`
	Hours string `xml:"hours,attr"`
	Monthdays string `xml:"monthdays,attr"`
	Weekdays string `xml:"weekdays,attr"`
	Yearsdays string `xml:"yearsdays,attr"`
	Months string `xml:"months,attr"`
	Weeks string `xml:"weeks,attr"`
	Years string `xml:"years,attr"`
	Weekyears string `xml:"weekyears,attr"`
	Moon string `xml:"moon,attr"`
}

type Rule struct {
	XMLName xml.Name
	Id *string `xml:"id,attr"`
	IdRef *string `xml:"id-ref,attr"`
	ScoreAttribute *string `xml:"score-attribute,attr"`
	BooleanOp *string `xml:"boolean-op,attr"`

	Attribute string `xml:"attribute,attr"`
	Operation string `xml:"operation,attr"`
	Value *string `xml:"value,attr"`
	Type *string `xml:"type,attr"`
	Start *string `xml:"start,attr"`
	End *string `xml:"end,attr"`
	Duration *DateSpec `xml:"duration"`
	DateSpec *DateSpec `xml:"date_spec"`

	Rules []Rule `xml:",any"`
}


// Named list of name-value pairs.
type NVSet struct {
	IdRef *string `xml:"id-ref,attr"`
	Id *string `xml:"id,attr"`
	Score *string `xml:"score,attr"`
	Rules []Rule `xml:"rule"`
	NVPairs []NVPair `xml:"nvpair"`
}

type Node struct {
	Id string `xml:"id,attr"`
	Uname string `xml:"uname,attr"`
	Type string `xml:"type,attr"`
	Description string `xml:"description,attr"`
	Score string `xml:"score,attr"`
	Attributes []NVSet `xml:"instance_attributes"`
	Utilization []NVSet `xml:"utilization"`
}

type Configuration struct {
	Options []NVSet `xml:"crm_config>cluster_property_set"`
	RscDefaults []NVSet `xml:"rsc_defaults>meta_attributes"`
	OpDefaults []NVSet `xml:"op_defaults>meta_attributes"`

	Nodes []Node `xml:"nodes>node"`

	Primitives []struct {
		Id string `xml:"id,attr"`
		Class string `xml:"class,attr"`
		Provider string `xml:"provider,attr"`
		Type string `xml:"type,attr"`
	} `xml:"resources>primitive"`

	Locations []struct {
		Id string `xml:"id,attr"`
	} `xml:"constraints>rsc_location"`
}

type LrmRscOp struct {
	Operation string `xml:"operation,attr"`
	CallId int `xml:"call-id,attr"`
	Rc int `xml:"rc-code,attr"`
	LastRun string `xml:"last-run,attr"`
	LastRcChange string `xml:"last-rc-change,attr"`
	ExecTime string `xml:"exec-time,attr"`
	QueueTime string `xml:"queue-time,attr"`
	OnNode string `xml:"on_node,attr"`
	ExitReason string `xml:"exit-reason,attr"`
}

type LrmResource struct {
	Id string `xml:"id,attr"`
	Type string `xml:"type,attr"`
	Class string `xml:"class,attr"`
	Provider string `xml:"provider,attr"`
	Ops []LrmRscOp `xml:"lrm_rsc_op"`
}

type NodeState struct {
	Id string `xml:"id,attr"`
	Uname string `xml:"uname,attr"`
	InCCM bool `xml:"in_ccm,attr"`
	Crmd string `xml:"crmd,attr"`
	CrmDebugOrigin string `xml:"crm-debug-origin,attr"`
	Join string `xml:"join,attr"`
	Expected string `xml:"expected,attr"`
	Resources []LrmResource `xml:"lrm>lrm_resources>lrm_resource"`
	Attributes []NVPair `xml:"transient_attributes>instance_attributes>nvpair"`
}

type Status struct {
	NodeState []NodeState `xml:"node_state"`
}

type CibVersion struct {
	AdminEpoch int32
	Epoch int32
	NumUpdates int32
}

func (ver *CibVersion) String() string {
	return fmt.Sprintf("%d:%d:%d", ver.AdminEpoch, ver.Epoch, ver.NumUpdates)
}

func NewCib() *Cib {
	var cib Cib
	cib.cCib = C.cib_new()
	return &cib
}

func (cib *Cib) SignOn(connection CibConnection) error {
	if cib.cCib.state == C.cib_connected_query || cib.cCib.state == C.cib_connected_command {
		return nil
	}

	rc := C.go_cib_signon(cib.cCib, C.crm_system_name, (uint32)(connection))
	if rc != C.pcmk_ok {
		return formatErrorRc((int)(rc))
	}
	return nil
}

func (cib *Cib) SignOff() error {
	rc := C.go_cib_signoff(cib.cCib)
	if rc != C.pcmk_ok {
		return formatErrorRc((int)(rc))
	}
	return nil
}

func (cib *Cib) Delete() {
	C.cib_delete(cib.cCib)
}

func (cib *Cib) queryImpl(xpath string, nochildren bool) (*C.xmlNode, error) {
	var root *C.xmlNode
	var rc C.int

	var opts C.int

	opts = C.cib_sync_call + C.cib_scope_local

	if xpath != "" {
		opts += C.cib_xpath
	}

	if nochildren {
		opts += C.cib_no_children
	}

	if xpath != "" {
		xp := C.CString(xpath)
		defer C.free(unsafe.Pointer(xp))
		rc = C.go_cib_query(cib.cCib, xp, (**C.xmlNode)(unsafe.Pointer(&root)), opts)
	} else {
		rc = C.go_cib_query(cib.cCib, nil, (**C.xmlNode)(unsafe.Pointer(&root)), opts)
	}
	if rc != C.pcmk_ok {
		return nil, formatErrorRc((int)(rc))
	}
	return root, nil
}


func (cib *Cib) Version() (*CibVersion, error) {
	var admin_epoch C.int
	var epoch C.int
	var num_updates C.int

	root, err := cib.queryImpl("/cib", true)
	if err != nil {
		return nil, err
	}
	defer C.free_xml(root)
	ok := C.cib_version_details(root, (*C.int)(unsafe.Pointer(&admin_epoch)), (*C.int)(unsafe.Pointer(&epoch)), (*C.int)(unsafe.Pointer(&num_updates)))
	if ok == 1 {
		return &CibVersion{(int32)(admin_epoch), (int32)(epoch), (int32)(num_updates)}, nil
	}
	return nil, &CibError{"Failed to get CIB version details"}
}

func (cib *Cib) Decode() error {
	text, err := cib.Query()
	if err != nil {
		return err
	}
	err = xml.Unmarshal(text, cib)
	if err != nil {
		return err
	}
	return nil
}

func (cib *Cib) Query() ([]byte, error) {
	var root *C.xmlNode
	root, err := cib.queryImpl("", false)
	if err != nil {
		return nil, err
	}
	defer C.free_xml(root)

	buffer := C.dump_xml_unformatted(root)
	defer C.free(unsafe.Pointer(buffer))

	return C.GoBytes(unsafe.Pointer(buffer), (C.int)(C.strlen(buffer))), nil
}

func (cib *Cib) QueryNoChildren() ([]byte, error) {
	var root *C.xmlNode
	root, err := cib.queryImpl("", true)
	if err != nil {
		return nil, err
	}
	defer C.free_xml(root)

	buffer := C.dump_xml_unformatted(root)
	defer C.free(unsafe.Pointer(buffer))

	return C.GoBytes(unsafe.Pointer(buffer), (C.int)(C.strlen(buffer))), nil
}


func (cib *Cib) QueryXPath(xpath string) ([]byte, error) {
	var root *C.xmlNode
	root, err := cib.queryImpl(xpath, false)
	if err != nil {
		return nil, err
	}
	defer C.free_xml(root)

	buffer := C.dump_xml_unformatted(root)
	defer C.free(unsafe.Pointer(buffer))

	return C.GoBytes(unsafe.Pointer(buffer), (C.int)(C.strlen(buffer))), nil
}

func (cib *Cib) QueryXPathNoChildren(xpath string) ([]byte, error) {
	var root *C.xmlNode
	root, err := cib.queryImpl(xpath, true)
	if err != nil {
		return nil, err
	}
	defer C.free_xml(root)

	buffer := C.dump_xml_unformatted(root)
	defer C.free(unsafe.Pointer(buffer))

	return C.GoBytes(unsafe.Pointer(buffer), (C.int)(C.strlen(buffer))), nil
}

func init() {
	s := C.CString("go-pacemaker")
	C.crm_log_init(s, C.LOG_CRIT, 0, 0, 0, nil, 1)
	C.free(unsafe.Pointer(s))
}

