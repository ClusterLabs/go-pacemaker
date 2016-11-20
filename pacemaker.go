// The pacemaker package provides an API for reading the Pacemaker cluster configuration (CIB).
package pacemaker

import (
	"unsafe"
	"fmt"
	"encoding/xml"
	"encoding/json"
	"bytes"
	"log"
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

	Attributes map[string]string `json:"attributes,omitempty"`

	Configuration Configuration `json:"configuration"`
	Status Status `json:"status"`
}

type RuleExpression struct {
	Id string `xml:"id,attr" json:"id"`
	Attribute string `xml:"attribute,attr" json:"attribute"`
	Operation string `xml:"operation,attr" json:"operation"`
	Value *string `xml:"value,attr,omitempty" json:"value,omitempty"`
	Type *string `xml:"type,attr,omitempty" json:"type,omitempty"`
}

type DateDuration struct {
	Id string `xml:"id,attr" json:"id"`
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

type DateExpression struct {
	Id string `xml:"id,attr" json:"id"`
	Operation *string `xml:"operation,attr,omitempty" json:"operation,omitempty"`
	Start *string `xml:"start,attr,omitempty" json:"start,omitempty"`
	End *string `xml:"end,attr,omitempty" json:"end,omitempty"`
	Duration *DateDuration `xml:"duration,omitempty" json:"duration,omitempty"`
	DateSpec *DateDuration `xml:"date_spec,omitempty" json:"date_spec,omitempty"`
}

type Rule struct {
	IdRef *string `json:"id-ref,omitempty"`
	Id *string `json:"id,omitempty"`
	Score *string `json:"score,omitempty"`
	ScoreAttribute *string `json:"score-attribute,omitempty"`
	BooleanOp *string `json:"boolean-op,omitempty"`
	Role *string `json:"role,omitempty"`
	Expressions []interface{} `json:"expressions,omitempty"`
}

type NVPair struct {
	Id *string `xml:"id,attr,omitempty" json:"id,omitempty"`
	IdRef *string `xml:"id-ref,attr,omitempty" json:"id-ref,omitempty"`
	Name *string `xml:"name,attr,omitempty" json:"name,omitempty"`
	Value *string `xml:"value,attr,omitempty" json:"value,omitempty"`
}

type AttributeSet struct {
	Id *string `json:"id,omitempty"`
	IdRef *string `json:"id-ref,omitempty"`
	Score *string `json:"score,omitempty"`
	Rules []Rule `json:"rules,omitempty"`
	Values []NVPair `json:"values,omitempty"`
}

type Op struct {
	Id string `xml:"id,attr" json:"id"`
	Name string `xml:"name,attr" json:"name"`
	Interval string `xml:"interval,attr" json:"interval"`
	Description *string `xml:"description,attr,omitempty" json:"description,omitempty"`
	StartDelay *string `xml:"start-delay,attr,omitempty" json:"start-delay,omitempty"`
	IntervalOrigin *string `xml:"interval-origin,attr,omitempty" json:"interval-origin,omitempty"`
	Timeout *string `xml:"timeout,attr,omitempty" json:"timeout,omitempty"`
	Enabled *bool `xml:"enabled,attr,omitempty" json:"enabled,omitempty"`
	RecordPending *bool `xml:"record-pending,attr,omitempty" json:"record-pending,omitempty"`
	Role *string `xml:"role,attr,omitempty" json:"role,omitempty"`
	Requires *string `xml:"requires,attr,omitempty" json:"requires,omitempty"`
	OnFail *string `xml:"on-fail,attr,omitempty" json:"on-fail,omitempty"`
	Attributes []AttributeSet `xml:"instance_attributes,omitempty" json:"attributes,omitempty"`
	Meta []AttributeSet `xml:"meta_attributes,omitempty" json:"meta,omitempty"`
}

type OpSet struct {
	Id *string `xml:"id,attr,omitempty" json:"id,omitempty"`
	IdRef *string `xml:"id-ref,attr,omitempty" json:"id-ref,omitempty"`
	Ops []Op `xml:"operation,omitempty" json:"ops,omitempty"`
}

type Resource struct {
	Id string `xml:"id,attr" json:"id"`
	Description *string `xml:"description,attr,omitempty" json:"description,omitempty"`
	Attributes []AttributeSet `xml:"instance_attributes,omitempty" json:"attributes,omitempty"`
	Meta []AttributeSet `xml:"meta_attributes,omitempty" json:"meta,omitempty"`
}


type Template struct {
	Resource
	Type string `xml:"type,attr" json:"type"`
	Class string `xml:"class,attr" json:"class"`
	Provider *string `xml:"provider,attr,omitempty" json:"provider,omitempty"`
	Utilization []AttributeSet `xml:"utilization,omitempty" json:"utilization,omitempty"`
	Ops []OpSet `xml:"operations,omitempty" json:"ops,omitempty"`
}


type Primitive struct {
	Resource
	Type *string `xml:"type,attr,omitempty" json:"type,omitempty"`
	Class *string `xml:"class,attr,omitempty" json:"class,omitemty"`
	Provider *string `xml:"provider,attr,omitempty" json:"provider,omitempty"`
	Template *string `xml:"template,attr,omitempty" json:"template,omitempty"`
	Utilization []AttributeSet `xml:"utilization,omitempty" json:"utilization,omitempty"`
	Ops []OpSet `xml:"operations,omitempty" json:"ops,omitempty"`
}


type Group struct {
	Resource
	Children []Primitive `xml:"primitive" json:"children"`
}

type Clone struct {
	Resource
	Child CibObject `json:"child"`
}

type Master struct {
	Resource
	Child CibObject `json:"child"`
}

type Constraint struct {
	Id string `xml:"id,attr" json:"id"`
}

type ResourceSet struct {
	Id *string `xml:"id,attr,omitempty" json:"id,omitempty"`
	IdRef *string `xml:"id-ref,attr,omitempty" json:"id-ref,omitempty"`
	Sequential *bool `json:"sequential,omitempty"`
	RequireAll *bool `json:"require-all,omitempty"`
	Ordering *string `json:"ordering,omitempty"`
	Action *string `json:"action,omitempty"`
	Role *string `json:"role,omitempty"`
	Score *string `json:"score,omitempty"`
	Resources []string `json:"resources,omitempty"`
}

type Location struct {
	Constraint
	Rsc *string `json:"rsc,omitempty"`
	RscPattern *string `json:"rsc-pattern,omitempty"`
	Role *string `json:"role,omitempty"`
	Score *string `json:"score,omitempty"`
	Node *string `json:"node,omitempty"`
	ResourceSets []ResourceSet `json:"resource-sets,omitempty"`
	ResourceDiscovery *string `json:"resource-discovery,omitempty"`
	Rules []Rule `json:"rules,omitempty"`
}

type Colocation struct {
	Constraint
	Score *string `json:"score,omitempty"`
	ScoreAttribute *string `json:"score-attribute,omitempty"`
	ScoreAttributeMangle *string `json:"score-attribute-mangle,omitempty"`
	ResourceSets []ResourceSet `json:"resource-sets,omitempty"`
	Rsc *string `json:"rsc,omitempty"`
	WithRsc *string `json:"with-rsc,omitempty"`
	NodeAttribute *string `json:"node-attribute,omitempty"`
	RscRole *string `json:"rsc-role,omitempty"`
	WithRscRole *string `json:"with-rsc-role,omitempty"`
}

type Order struct {
	Constraint
	Symmetrical *bool `json:"symmetrical,omitempty"`
	RequireAll *bool `json:"require-all,omitempty"`
	Score *string `json:"score,omitempty"`
	Kind *string `json:"kind,omitempty"`
	ResourceSets []ResourceSet `json:"resource-sets,omitempty"`
	First *string `json:"first,omitempty"`
	Then *string `json:"then,omitempty"`
	FirstAction *string `json:"first-action,omitempty"`
	ThenAction *string `json:"then-action,omitempty"`
}

type Ticket struct {
	Constraint
	ResourceSets []ResourceSet `json:"resource-sets,omitempty"`
	Rsc *string `json:"rsc,omitempty"`
	RscRole *string `json:"rsc-role,omitempty"`
	Ticket string `json:"ticket"`
	LossPolicy *string `json:"loss-policy,omitempty"`
}

type Node struct {
	Id string `json:"id"`
	Uname string `json:"uname"`
	Type *string `json:"type,omitempty"`
	Description *string `json:"description,omitempty"`
	Score *string `json:"score,omitempty"`
	Attributes []AttributeSet `json:"attributes,omitempty"`
	Utilization []AttributeSet `json:"utilization,omitempty"`
}

type FencingLevel struct {
	Id string `json:"id"`
	Target *string `json:"target,omitempty"`
	TargetPattern *string `json:"target-pattern,omitempty"`
	TargetAttribute *string `json:"target-attribute,omitempty"`
	TargetValue *string `json:"target-value,omitempty"`
	Index int `json:"index"`
	Devices string `json:"devices"`
}

type AclTarget struct {
	Id string `json:"id"`
	Roles []string `json:"roles,omitempty"`
}

type AclPermission struct {
	Id string `json:"id"`
	Kind string `json:"kind"`
	Xpath *string `json:"xpath,omitempty"`
	Reference *string `json:"reference,omitempty"`
	ObjectType *string `json:"object-type,omitempty"`
	Attribute *string `json:"attribute,omitempty"`
	Description *string `json:"description,omitempty"`
}

type AclRole struct {
	Id string `json:"id"`
	Description *string `json:"description,omitempty"`
	Permissions []AclPermission `json:"permissions,omitempty"`
}

type Tag struct {
	Id string `json:"id"`
	References []string `json:"references,omitempty"`
}

type Recipient struct {
	Id string `json:"id"`
	Description *string `json:"description,omitempty"`
	Value string `json:"value"`
	Meta []AttributeSet `json:"meta,omitempty"`
	Attributes []AttributeSet `json:"attributes,omitempty"`
}

type Alert struct {
	Id string `xml:"id,attr,omitempty" json:"id"`
	Description *string `xml:"description,attr,omitempty" json:"description,omitempty"`
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
	Resources []CibObject `json:"resources,omitempty"`
	Constraints []CibObject `json:"constraints,omitempty"`
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


func decodeTagImpl(decoder *xml.Decoder, se *xml.StartElement,
	attrfn func(attr xml.Attr) bool, elemfn  func(decoder *xml.Decoder, parent *xml.StartElement, current *xml.StartElement, depth int) bool) {
	depth := 1
	var ret bool
	for _, attr := range se.Attr {
		ret = attrfn(attr)
		if ret {
			return
		}
	}
	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		switch ce := t.(type) {
		case xml.StartElement:
			if ce.Name.Local == se.Name.Local {
				depth++
			}
			ret = elemfn(decoder, se, &ce, depth)
			if ret {
				return
			}
		case xml.EndElement:
			if ce.Name.Local == se.Name.Local {
				depth--
				if depth == 0 {
					return
				}
			}
		}
	}
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
			} else if (se.Name.Local == "configuration") {
				decodeConfiguration(decoder, &cib.Configuration)
			} else if (se.Name.Local == "status") {
				decoder.DecodeElement(&cib.Status, &se)
			}
		}
	}
	return nil
}

func decodeNVPair(decoder *xml.Decoder, se *xml.StartElement) *NVPair {
	a := &NVPair{}
	decodeTagImpl(decoder, se, func(attr xml.Attr) bool {
		switch attr.Name.Local {
		case "id":
			a.Id = &attr.Value
		case "id-ref":
			a.IdRef = &attr.Value
		case "name":
			a.Name = &attr.Value
		case "value":
			a.Value = &attr.Value
		default:
			log.Printf("Warning: Unknown nvpair attribute '%s'", attr.Name.Local)
		}
		return true
	}, nil)
	return a
}

func decodeRule(decoder *xml.Decoder, se *xml.StartElement) *Rule {
	a := &Rule{}
	decodeTagImpl(decoder, se, func(attr xml.Attr) bool {
		switch attr.Name.Local {
		case "id":
			a.Id = &attr.Value
		case "id-ref":
			a.IdRef = &attr.Value
		case "score":
			a.Score = &attr.Value
		case "score-attribute":
			a.ScoreAttribute = &attr.Value
		case "boolean-op":
			a.BooleanOp = &attr.Value
		case "role":
			a.Role = &attr.Value
		default:
			log.Printf("Warning: Unknown nvpair attribute '%s'", attr.Name.Local)
		}
		return false
	}, func(decoder *xml.Decoder, parent *xml.StartElement, current *xml.StartElement, depth int) bool {
		if current.Name.Local == "expression" {
			n := decodeRuleExpression(decoder, current)
			if n != nil {
				a.Expressions = append(a.Expressions, *n)
			}
		} else if current.Name.Local == "date_expression" {
			n := decodeDateExpression(decoder, current)
			if n != nil {
				a.Expressions = append(a.Expressions, *n)
			}
		} else if current.Name.Local == "rule" {
			r := decodeRule(decoder, current)
			if r != nil {
				a.Expressions = append(a.Expressions, *r)
			}
		}
		return false
	})
	return a
}

func decodeAttributeSet(decoder *xml.Decoder, se *xml.StartElement) *AttributeSet {
	a := &AttributeSet{}
	depth := 1
	for _, attr := range se.Attr {
		switch attr.Name.Local {
		case "id":
			a.Id = &attr.Value
		case "id-ref":
			a.IdRef = &attr.Value
		case "score":
			a.Score = &attr.Value
		default:
			log.Printf("Warning: Unknown nvset attribute '%s'", attr.Name.Local)
		}
	}
	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		switch ce := t.(type) {
		case xml.StartElement:
			if ce.Name.Local == se.Name.Local {
				depth++
			} else if depth == 1 {
				if ce.Name.Local == "nvpair" {
					n := decodeNVPair(decoder, &ce)
					if n != nil {
						a.Values = append(a.Values, *n)
					}
				} else if ce.Name.Local == "rule" {
					r := decodeRule(decoder, &ce)
					if r != nil {
						a.Rules = append(a.Rules, *r)
					}
				}
			}
		case xml.EndElement:
			if ce.Name.Local == se.Name.Local {
				depth--
				if depth == 0 {
					return a
				}
			}
		}
	}
	return a
}


func decodeConfiguration(decoder *xml.Decoder, configuration *Configuration) {
	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		switch se := t.(type) {
		case xml.StartElement:
			switch se.Name.Local {
			case "cluster_property_set":
				a := decodeAttributeSet(decoder, &se)
				if a != nil {
					configuration.CrmConfig = append(configuration.CrmConfig, *a)
				}
			case "rsc_defaults":
				a := decodeAttributeSet(decoder, &se)
				if a != nil {
					configuration.RscDefaults = append(configuration.RscDefaults, *a)
				}
			case "op_defaults":
				a := decodeAttributeSet(decoder, &se)
				if a != nil {
					configuration.OpDefaults = append(configuration.OpDefaults, *a)
				}
			case "nodes":
			case "resources":
			case "constraints":
			case "acls":
			case "tags":
			case "alerts":
			case "fencing-topology":
			}
		case xml.EndElement:
			if se.Name.Local == "configuration" {
				return
			}
		}
	}
}
