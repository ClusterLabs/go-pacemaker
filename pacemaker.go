// The pacemaker package provides an API for reading the Pacemaker cluster configuration (CIB).
package pacemaker

import (
	"unsafe"
	"fmt"
	"encoding/xml"
	"encoding/json"
	"bytes"
//	"log"
//	"reflect"
	"strings"
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
	Attributes map[string]string `pcmk:"cib-attrs" json:"attributes,omitempty"`
	Configuration Configuration `xml:"configuration" json:"configuration"`
	Status Status `xml:"status" json:"status"`
}

type idMixin struct {
	Id string `xml:"id,attr" json:"id"`
}

type idRefMixin struct {
	IdRef *string `xml:"id-ref,attr,omitempty" json:"id-ref,omitempty"`
	Id *string `xml:"id,attr,omitempty" json:"id,omitempty"`
}

type scoreMixin struct {
	Score *string `xml:"score,attr,omitempty" json:"score,omitempty"`
}

type descMixin struct {
	Description *string `xml:"description,attr,omitempty" json:"description,omitempty"`
}

// expression
// can be marshalled
type RuleExpression struct {
	idMixin
	Attribute string `xml:"attribute,attr" json:"attribute"`
	Operation string `xml:"operation,attr" json:"operation"`
	Value *string `xml:"value,attr,omitempty" json:"value,omitempty"`
	Type *string `xml:"type,attr,omitempty" json:"type,omitempty"`
}

// date_spec and duration are both backed by this type
// can be marshalled
type DateDuration struct {
	idMixin
	Hours *string `xml:"hours,attr,omitempty" json:"hours,omitempty"`
	Monthdays *string `xml:"monthdays,attr,omitempty" json:"monthdays,omitempty"`
	Weekdays *string `xml:"weekdays,attr,omitempty" json:"weekdays,omitempty"`
	Yearsdays *string `xml:"yearsdays,attr,omitempty" json:"yearsdays,omitempty"`
	Months *string `xml:"months,attr,omitempty" json:"months,omitempty"`
	Weeks *string `xml:"weeks,attr,omitempty" json:"weeks,omitempty"`
	Years *string `xml:"years,attr,omitempty" json:"years,omitempty"`
	Weekyears *string `xml:"weekyears,attr,omitempty" json:"weekyears,omitempty"`
	Moon *string `xml:"moon,attr,omitempty" json:"moon,omitempty"`
}

// can be marshalled
type DateExpression struct {
	idMixin
	Operation *string `xml:"operation,attr,omitempty" json:"operation,omitempty"`
	Start *string `xml:"start,attr,omitempty" json:"start,omitempty"`
	End *string `xml:"end,attr,omitempty" json:"end,omitempty"`
	Duration *DateDuration `xml:"duration,omitempty" json:"duration,omitempty"`
	DateSpec *DateDuration `xml:"date_spec,omitempty" json:"date_spec,omitempty"`
}

// can NOT be marshalled
type Rule struct {
	idRefMixin
	scoreMixin
	ScoreAttribute *string `xml:"score-attribute,attr,omitempty" json:"score-attribute,omitempty"`
	BooleanOp *string `xml:"boolean-op,attr,omitempty" json:"boolean-op,omitempty"`
	Role *string `xml:"role,attr,omitempty" json:"role,omitempty" validate:"ValidAttributeRole"`
	Expressions []interface{} `pcmk:"rule-expression" xml:"expression,omitempty" json:"expressions,omitempty"`
}

// can be marshalled
type NVPair struct {
	idRefMixin
	Name *string `xml:"name,attr,omitempty" json:"name,omitempty"`
	Value *string `xml:"value,attr,omitempty" json:"value,omitempty"`
}

// can NOT be marshalled (due to rules)
type AttributeSet struct {
	idRefMixin
	scoreMixin
	Rules []Rule `xml:"rule,omitempty" json:"rules,omitempty"`
	Values []NVPair `xml:"nvpair,omitempty" json:"values,omitempty"`
}

// can NOT be marshalled (due to attributesets)
type Op struct {
	idMixin
	descMixin
	Name string `xml:"name,attr" json:"name"`
	Interval string `xml:"interval,attr" json:"interval"`
	StartDelay *string `xml:"start-delay,attr,omitempty" json:"start-delay,omitempty"`
	IntervalOrigin *string `xml:"interval-origin,attr,omitempty" json:"interval-origin,omitempty"`
	Timeout *string `xml:"timeout,attr,omitempty" json:"timeout,omitempty"`
	Enabled *bool `xml:"enabled,attr,omitempty" json:"enabled,omitempty"`
	RecordPending *bool `xml:"record-pending,attr,omitempty" json:"record-pending,omitempty"`
	Role *string `xml:"role,attr,omitempty" json:"role,omitempty" validate:"ValidAttributeRole"`
	Requires *string `xml:"requires,attr,omitempty" json:"requires,omitempty" validate:"ValidOperationRequires"`
	OnFail *string `xml:"on-fail,attr,omitempty" json:"on-fail,omitempty" validate:"ValidOperationFail"`
	Attributes []AttributeSet `xml:"instance_attributes,omitempty" json:"attributes,omitempty"`
	Meta []AttributeSet `xml:"meta_attributes,omitempty" json:"meta,omitempty"`
}

// can NOT be marshalled (due to op/attributesets)
type OpSet struct {
	idRefMixin
	Ops []Op `xml:"operation,omitempty" json:"ops,omitempty"`
}

// can NOT be marshalled
type Resource struct {
	idMixin
	descMixin
	Attributes []AttributeSet `xml:"instance_attributes,omitempty" json:"attributes,omitempty"`
	Meta []AttributeSet `xml:"meta_attributes,omitempty" json:"meta,omitempty"`
}


// can NOT be marshalled
type Template struct {
	Resource
	Type string `xml:"type,attr" json:"type"`
	Class string `xml:"class,attr" json:"class" validate:"ValidResourceClass"`
	Provider *string `xml:"provider,attr,omitempty" json:"provider,omitempty"`
	Utilization []AttributeSet `xml:"utilization,omitempty" json:"utilization,omitempty"`
	Ops []OpSet `xml:"operations,omitempty" json:"ops,omitempty"`
}


// can NOT be marshalled
type Primitive struct {
	Resource
	Type *string `xml:"type,attr,omitempty" json:"type,omitempty"`
	Class *string `xml:"class,attr,omitempty" json:"class,omitemty" validate:"ValidResourceClass"`
	Provider *string `xml:"provider,attr,omitempty" json:"provider,omitempty"`
	Template *string `xml:"template,attr,omitempty" json:"template,omitempty"`
	Utilization []AttributeSet `xml:"utilization,omitempty" json:"utilization,omitempty"`
	Ops []OpSet `xml:"operations,omitempty" json:"ops,omitempty"`
}


// can NOT be marshalled
type Group struct {
	Resource
	Children []Primitive `xml:"primitive" json:"children"`
}

// can NOT be marshalled
type Clone struct {
	Resource
	Child CibObject `xml:"child" json:"child"`
}

// can NOT be marshalled
type Master struct {
	Resource
	Child CibObject `xml:"child" json:"child"`
}

// can NOT be marshalled
type Constraint struct {
	idMixin
}

// can NOT be marshalled
type ResourceSet struct {
	idRefMixin
	Sequential *bool `xml:"sequential,attr,omitempty" json:"sequential,omitempty"`
	RequireAll *bool `xml:"require-all,attr,omitempty" json:"require-all,omitempty"`
	Ordering *string `xml:"ordering,attr,omitempty" json:"ordering,omitempty" validate:"ValidConstraintOrdering"`
	Action *string `xml:"action,attr,omitempty" json:"action,omitempty" validate:"ValidAttributeAction"`
	Role *string `xml:"role,attr,omitempty" json:"role,omitempty" validate:"ValidAttributeRole"`
	scoreMixin
	Resources []string `xml:"resource_ref,omitempty" json:"resources,omitempty"`
}

// can NOT be marshalled
type Location struct {
	Constraint
	Rsc *string `xml:"rsc,attr,omitempty" json:"rsc,omitempty"`
	RscPattern *string `xml:"rsc-pattern,attr,omitempty" json:"rsc-pattern,omitempty"`
	Role *string `xml:"role,attr,omitempty" json:"role,omitempty"`
	scoreMixin
	Node *string `xml:"node,attr,omitempty" json:"node,omitempty"`
	ResourceSets []ResourceSet `xml:"resource-set,omitempty" json:"resource-sets,omitempty"`
	ResourceDiscovery *string `xml:"resource-discovery,attr,omitempty" json:"resource-discovery,omitempty" validate:"ValidAttributeDiscovery"`
	Rules []Rule `xml:"rule,omitempty" json:"rules,omitempty"`
}

// can NOT be marshalled
type Colocation struct {
	Constraint
	scoreMixin
	ResourceSets []ResourceSet `xml:"resource-set" json:"resource-sets,omitempty"`
	Rsc *string `xml:"rsc,attr,omitempty" json:"rsc,omitempty"`
	WithRsc *string `xml:"with-rsc,attr,omitempty" json:"with-rsc,omitempty"`
	NodeAttribute *string `xml:"node-attribute,attr,omitempty" json:"node-attribute,omitempty"`
	RscRole *string `xml:"rsc-role,attr,omitempty" json:"rsc-role,omitempty"`
	WithRscRole *string `xml:"with-rsc-role,attr,omitempty" json:"with-rsc-role,omitempty"`
}

// can NOT be marshalled
type Order struct {
	Constraint
	Symmetrical *bool `xml:"symmetrical,attr,omitempty" json:"symmetrical,omitempty"`
	RequireAll *bool `xml:"require-all,attr,omitempty" json:"require-all,omitempty"`
	scoreMixin
	Kind *string `xml:"kind,attr,omitempty" json:"kind,omitempty"`
	ResourceSets []ResourceSet `xml:"resource-set" json:"resource-sets,omitempty"`
	First *string `xml:"first,attr,omitempty" json:"first,omitempty"`
	Then *string `xml:"then,attr,omitempty" json:"then,omitempty"`
	FirstAction *string `xml:"first-action,attr,omitempty" json:"first-action,omitempty" validate:"ValidAttributeAction"`
	ThenAction *string `xml:"then-action,attr,omitempty" json:"then-action,omitempty" validate:"ValidAttributeAction"`
}

// can NOT be marshalled
type Ticket struct {
	Constraint
	ResourceSets []ResourceSet `xml:"resource-set" json:"resource-sets,omitempty"`
	Rsc *string `xml:"rsc,attr,omitempty" json:"rsc,omitempty"`
	RscRole *string `xml:"rsc-role,attr,omitempty" json:"rsc-role,omitempty" validate:"ValidAttributeRole"`
	Ticket string `xml:"ticket,attr" json:"ticket"`
	LossPolicy *string `xml:"loss-policy,attr,omitempty" json:"loss-policy,omitempty" validate:"ValidTicketLossPolicy"`
}

// can NOT be marshalled
type Node struct {
	idMixin
	descMixin
	scoreMixin
	Uname string `xml:"uname,attr" json:"uname"`
	Type *string `xml:"type,attr,omitempty" json:"type,omitempty" validate:"ValidNodeType"`
	Attributes []AttributeSet `xml:"instance_attributes" json:"attributes,omitempty"`
	Utilization []AttributeSet `xml:"utilization" json:"utilization,omitempty"`
}

// can be marshalled
type FencingLevel struct {
	idMixin
	Target *string `xml:"target,attr,omitempty" json:"target,omitempty"`
	TargetPattern *string `xml:"target-pattern,attr,omitempty" json:"target-pattern,omitempty"`
	TargetAttribute *string `xml:"target-attribute,attr,omitempty" json:"target-attribute,omitempty"`
	TargetValue *string `xml:"target-value,attr,omitempty" json:"target-value,omitempty"`
	Index int `xml:"index,attr" json:"index"`
	Devices string `xml:"devices,attr" json:"devices"`
}

// can be marshalled
type AclTarget struct {
	idMixin
	Roles []string `xml:"role" json:"roles,omitempty"`
}

// can be marshalled
type AclPermission struct {
	idMixin
	descMixin
	Kind string `xml:"kind,attr" json:"kind" validate:"ValidPermissionKind"`
	Xpath *string `xml:"xpath,attr,omitempty" json:"xpath,omitempty"`
	Reference *string `xml:"reference,attr,omitempty" json:"reference,omitempty"`
	ObjectType *string `xml:"object-type,attr,omitempty" json:"object-type,omitempty"`
	Attribute *string `xml:"attribute,attr,omitempty" json:"attribute,omitempty"`
}

// can be marshalled
type AclRole struct {
	idMixin
	descMixin
	Permissions []AclPermission `xml:"acl_permission" json:"permissions,omitempty"`
}

// can NOT be marshalled
type Tag struct {
	idMixin
	References []string `xml:"obj_ref" json:"references,omitempty"`
}

// can NOT be marshalled
type Recipient struct {
	idMixin
	descMixin
	Value string `xml:"value,attr" json:"value"`
	Meta []AttributeSet `xml:"meta_attributes" json:"meta,omitempty"`
	Attributes []AttributeSet `xml:"instance_attributes" json:"attributes,omitempty"`
}

// can NOT be marshalled
type Alert struct {
	idMixin
	descMixin
	Path string `xml:"path,attr" json:"path"`
	Meta []AttributeSet `xml:"meta_attributes,omitempty" json:"meta,omitempty"`
	Attributes []AttributeSet `xml:"instance_attributes,omitempty" json:"attributes,omitempty"`
	Recipients []Recipient `xml:"recipient,omitempty" json:"recipients,omitempty"`
}

// /cib/configuration
// needs custom parse + encode postprocess to recreate missing ids
type Configuration struct {
	CrmConfig []AttributeSet `xml:"crm_config>cluster_property_set,omitempty" json:"crm_config,omitempty"`
	RscDefaults []AttributeSet `xml:"rsc_defaults>meta_attributes,omitempty" json:"rsc_defaults,omitempty"`
	OpDefaults []AttributeSet `xml:"op_defaults>meta_attributes,omitempty" json:"op_defaults,omitempty"`
	Nodes []Node `xml:"nodes>node,omitempty" json:"nodes,omitempty"`
	Resources []CibObject `xml:"resource" json:"resources,omitempty"`
	Constraints []CibObject `xml:"constraint" json:"constraints,omitempty"`
	Fencing []FencingLevel `xml:"fencing-topology>fencing-level,omitempty" json:"fencing-topology,omitempty"`
	AclTargets []AclTarget `xml:"acls>acl_target,omitempty" json:"acl-targets,omitempty"`
	AclRoles []AclRole `xml:"acls>acl_role,omitempty" json:"acl-roles,omitempty"`
	Tags []Tag `xml:"tags>tag,omitempty" json:"tags,omitempty"`
	Alerts []Alert `xml:"alerts>alert,omitempty" json:"alerts,omitempty"`
}

// /cib/status/node_state/lrm/lrm_resources/lrm_resource/lrm_rsc_op
// Can be marshalled/unmarshalled
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

// /cib/status/node_state/lrm/lrm_resources/lrm_resource
// Can be marshalled/unmarshalled
type ResourceState struct {
	Id string `xml:"id,attr" json:"id,omitempty"`
	Type string `xml:"type,attr" json:"type,omitempty"`
	Class string `xml:"class,attr" json:"class,omitempty"`
	Provider string `xml:"provider,attr" json:"provider,omitempty"`
	Ops []ResourceStateOp `xml:"lrm_rsc_op" json:"ops,omitempty"`
}

// /cib/status/node_state/transient_attributes/instance_attributes/nvpair
// Can be marshalled/unmarshalled
type SimpleNVPair struct {
	Name string `xml:"name,attr" json:"name"`
	Value string `xml:"value,attr" json:"value"`
}

// /cib/status/node_state
// Can be marshalled/unmarshalled
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

// /cib/status
// Can be marshalled/unmarshalled
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
	if err = cib.decodeCibObjects(xmldata); err != nil {
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


func stringToBool(bstr string) bool {
	sl := strings.ToLower(bstr)
	return sl == "true" || sl == "on" || sl == "yes" || sl == "y" || sl == "1"
}


// Read XML configuration into an object tree.
// To save, we want a series of crmsh commands
// so no need for objects -> xml serialization
// at least. Just save a list of operations
// performed and apply them all on a shadow cib.
func (cib *Cib) decodeCibObjects(xmldata []byte) error {
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
