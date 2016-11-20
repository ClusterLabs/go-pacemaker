// The pacemaker package provides an API for reading the Pacemaker cluster configuration (CIB).
package pacemaker

import (
	"unsafe"
	"fmt"
	"encoding/xml"
	"encoding/json"
	"bytes"
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

type CibObject struct {
}

type CibSerialize interface {
	FromXml([]byte) CibObject
	FromJson([]byte) CibObject
	ToJson() []byte
	ToXml() []byte
}

// Root entity representing the CIB. Can be
// populated with CIB data if the Decode
// method is used.
type Cib struct {
	cCib *C.cib_t

	Attributes map[string]string

	Configuration Configuration
	Status Status
}

type RuleExpression struct {
	Id string
	Attribute string
	Operation string
	Value *string
	Type *string
}

type DateDuration struct {
	Id string
	Hours *string
	Monthdays *string
	Weekdays *string
	Yearsdays *string
	Months *string
	Weeks *string
	Years *string
	Weekyears *string
	Moon *string
}

type DateExpression struct {
	Id string
	Operation *string
	Start *string
	End *string
	Duration *DateDuration
}

type Rule struct {
	IdRef *string
	Id *string
	Score *string
	ScoreAttribute *string
	BooleanOp *string
	Role *string
	Expressions []interface{}
}

type NVPair struct {
	Id *string
	IdRef *string
	Name *string
	Value *string
}

type AttributeSet struct {
	Id *string
	IdRef *string
	Score *string
	Rules []Rule
	Values []NVPair
}

type Operation struct {
	Id string
	Name string
	Interval string
	Description *string
	StartDelay *string
	IntervalOrigin *string
	Timeout *string
	Enabled *bool
	RecordPending *bool
	Role *string
	Requires *string
	OnFail *string
	Meta []AttributeSet
	Attributes []AttributeSet
}

type OperationSet struct {
	Id *string
	IdRef *string
	Operations []Operation
}

type Resource struct {
	Id string
	Description *string
	Attributes []AttributeSet
	Meta []AttributeSet
}

type Primitive struct {
	Resource
	Type *string
	Class *string
	Provider *string
	Template *string
	Utilization []AttributeSet
	Operations []OperationSet
}

type Template struct {
	Resource
	Type string
	Class string
	Provider *string
	Utilization []AttributeSet
	Operations []OperationSet
}


type Group struct {
	Resource
	Children []Primitive
}

type Clone struct {
	Resource
	Child CibObject
}

type Master struct {
	Resource
	Child CibObject
}

type Constraint struct {
	Id string
}

type ResourceSet struct {
	Id *string
	IdRef *string
	Sequential *bool
	RequireAll *bool
	Ordering *string
	Action *string
	Role *string
	Score *string
	Resources []string
}

type Location struct {
	Constraint
	Rsc *string
	RscPattern *string
	Role *string
	Score *string
	Node *string
	ResourceSets []ResourceSet
	ResourceDiscovery *string
	Rules []Rule
}

type Colocation struct {
	Constraint
	Score *string
	ScoreAttribute *string
	ScoreAttributeMangle *string
	ResourceSets []ResourceSet
	Rsc *string
	WithRsc *string
	NodeAttribute *string
	RscRole *string
	WithRscRole *string
}

type Order struct {
	Constraint
	Symmetrical *bool
	RequireAll *bool
	Score *string
	Kind *string
	ResourceSets []ResourceSet
	First *string
	Then *string
	FirstAction *string
	ThenAction *string
}

type Ticket struct {
	Constraint
	ResourceSets []ResourceSet
	Rsc *string
	RscRole *string
	Ticket string
	LossPolicy *string
}

type Node struct {
	Id string
	Uname string
	Type *string
	Description *string
	Score *string
	Attributes []AttributeSet
	Utilization []AttributeSet
}

type FencingLevel struct {
	Id string
	Target *string
	TargetPattern *string
	TargetAttribute *string
	TargetValue *string
	Index int
	Devices string
}

type AclTarget struct {
	Id string
	Roles []string
}

type AclPermission struct {
	Id string
	Kind string
	Xpath *string
	Reference *string
	ObjectType *string
	Attribute *string
	Description *string
}

type AclRole struct {
	Id string
	Description *string
	Permissions []AclPermission
}

type Tag struct {
	Id string
	References []string
}

type Recipient struct {
	Id string
	Description *string
	Value string
	Meta []AttributeSet
	Attributes []AttributeSet
}

type Alert struct {
	Id string
	Description *string
	Path string
	Meta []AttributeSet
	Attributes []AttributeSet
	Recipients []Recipient
}

type Configuration struct {
	CrmConfig []AttributeSet
	RscDefaults []AttributeSet
	OpDefaults []AttributeSet
	Nodes []Node
	Resources []CibObject
	Constraints []CibObject
	Fencing []FencingLevel
	AclTargets []AclTarget
	AclRoles []AclRole
	Tags []Tag
	Alerts []Alert
}

type ResourceStateOp struct {
	Operation string `xml:"operation,attr" json:"operation,omitempty"`
	CallId int `xml:"call-id,attr" json:"call-id,omitempty"`
	Rc int `xml:"rc-code,attr" json:"rc-code,omitempty"`
	LastRun string `xml:"last-run,attr" json:"last-run,omitempty"`
	LastRcChange string `xml:"last-rc-change,attr" json:"last-rc-change,omitempty"`
	ExecTime string `xml:"exec-time,attr" json:"exec-time,omitempty"`
	QueueTime string `xml:"queue-time,attr" json:"queue-time,omitempty"`
	OnNode string `xml:"on_node,attr" json:"on-node,omitempty"`
	ExitReason string `xml:"exit-reason,attr" json:"exit-reason,omitempty"`
}

type ResourceState struct {
	Id string `xml:"id,attr" json:"id,omitempty"`
	Type string `xml:"type,attr" json:"type,omitempty"`
	Class string `xml:"class,attr" json:"class,omitempty"`
	Provider string `xml:"provider,attr" json:"provider,omitempty"`
	Ops []ResourceStateOp `xml:"lrm_rsc_op" json:"ops,omitempty"`
}

type SimpleNVPair struct {
	Name string `xml:"name,attr" json:"name"`
	Value string `xml:"value,attr" json:"value"`
}

type NodeState struct {
	Id string `xml:"id,attr" json:"id,omitempty"`
	Uname string `xml:"uname,attr" json:"uname,omitempty"`
	InCcm bool `xml:"in_ccm,attr" json:"in-ccm,omitempty"`
	Crmd string `xml:"crmd,attr" json:"crmd,omitempty"`
	CrmDebugOrigin string `xml:"crm-debug-origin,attr" json:"crm-debug-origin,omitempty"`
	Join string `xml:"join,attr" json:"join,omitempty"`
	Expected string `xml:"expected,attr" json:"expected,omitempty"`
	Resources []ResourceState `xml:"lrm>lrm_resources>lrm_resource" json:"resources,omitempty"`
	Attributes []SimpleNVPair `xml:"transient_attributes>instance_attributes>nvpair" json:"attributes,omitempty"`
}

type Status struct {
	NodeState []NodeState `xml:"node_state" json:"node-state"`
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

func NewCibFile(filename string) *Cib {
	var cib Cib
	s := C.CString(filename)
	cib.cCib = C.cib_file_new(s)
	C.free(unsafe.Pointer(s))
	return &cib
}

func NewCibShadow(name string) *Cib {
	var cib Cib
	s := C.CString(name)
	cib.cCib = C.cib_shadow_new(s)
	C.free(unsafe.Pointer(s))
	return &cib
}

func GetShadowFile(name string) string {
	s := C.CString(name)
	defer C.free(unsafe.Pointer(s))
	return C.GoString(C.get_shadow_file(s))
}

func NewCibRemote(server, user, passwd string, port int, encrypted bool) *Cib {
	var cib Cib
	s := C.CString(server)
	u := C.CString(user)
	p := C.CString(passwd)
	var e C.int = 0
	if encrypted {
		e = 1
	}
	cib.cCib = C.cib_remote_new(s, u, p, (C.int)(port), (C.gboolean)(e))
	C.free(unsafe.Pointer(s))
	C.free(unsafe.Pointer(u))
	C.free(unsafe.Pointer(p))
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
	xmldata, err := cib.Query()
	if err != nil {
		return err
	}
	if err = cib.loadCibObjects(xmldata); err != nil {
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

func (status *Status) ToJson() ([]byte, error) {
	return json.Marshal(status)
}

func init() {
	s := C.CString("go-pacemaker")
	C.crm_log_init(s, C.LOG_CRIT, 0, 0, 0, nil, 1)
	C.free(unsafe.Pointer(s))
}

var ValidOperationRole = []string{"Stopped", "Started", "Slave", "Master"}
var ValidOperationRequires = []string{"nothing", "quorum", "fencing", "unfencing"}
var ValidOperationOnFail = []string{"ignore", "block", "stop", "restart", "standby", "fence", "restart-container"}
var ValidResourceClass = []string{"ocf", "lsb", "heartbeat", "stonith", "upstart", "service", "systemd", "nagios"}
var ValidConstraintOrdering = []string{"group", "listed"}
var ValidTicketLossPolicy = []string{"stop", "demote", "fence", "freeze"}
var ValidAttributeDiscovery = []string{"always", "never", "exclusive"}
var ValidAttributeAction = []string{"start", "promote", "demote", "stop"}
var ValidAttributeRole = []string{"Stopped", "Started", "Master", "Slave"}
var ValidOrderType = []string{"Optional", "Mandatory", "Serialize"}
var ValidPermissionKind = []string{"read", "write", "deny"}
var ValidNodeType = []string{"normal", "member", "ping", "remote"}

// Read XML configuration into an object tree.
// To save, we want a series of crmsh commands
// so no need for objects -> xml serialization
// at least. Just save a list of operations
// performed and apply them all on a shadow cib.
func (cib *Cib) loadCibObjects(xmldata []byte) error {
	decoder := xml.NewDecoder(bytes.NewReader(xmldata))
	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "cib" {
				cib.Attributes = make(map[string]string)
				for _, attr := range se.Attr {
					cib.Attributes[attr.Name.Local] = attr.Value
				}
			} else if (se.Name.Local == "status") {
				decoder.DecodeElement(&cib.Status, &se)
			}
		}
	}
	return nil
}
