// The pacemaker package provides an API for reading the Pacemaker cluster configuration (CIB).
package pacemaker

import (
	"unsafe"
	"fmt"
	"encoding/xml"
	"encoding/json"
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

type CibObject interface {
	FromJson([]byte) (error)
	ToJson() ([]byte, error)
}

// Root entity representing the CIB. Can be
// populated with CIB data if the Decode
// method is used.
type Cib struct {
	cCib *C.cib_t
	XMLName xml.Name `xml:"cib" json:"-"`
	ValidateWith *string `xml:"validate-with,attr,omitempty" json:"validate-with,omitempty"`
	AdminEpoch *string `xml:"admin_epoch,attr,omitempty" json:"admin_epoch,omitempty"`
	Epoch *string `xml:"epoch,attr,omitempty" json:"epoch,omitempty"`
	NumUpdates *string `xml:"num_updates,attr,omitempty" json:"num_updates,omitempty"`
	CrmFeatureSet *string `xml:"crm_feature_set,attr,omitempty" json:"crm_feature_set,omitempty"`
	RemoteTLSPort *string `xml:"remote-tls-port,attr,omitempty" json:"remote-tls-port,omitempty"`
	RemoteClearPort *string `xml:"remote-clear-port,attr,omitempty" json:"remote-clear-port,omitempty"`
	HaveQuorum *string `xml:"have-quorum,attr,omitempty" json:"have-quorum,attr,omitempty"`
	DcUuid *string `xml:"dc-uuid,attr,omitempty" json:"dc-uuid,attr,omitempty"`
	CibLastWritten *string `xml:"cib-last-written,attr,omitempty" json:"cib-last-written,attr,omitempty"`
	NoQuorumPanic *string `xml:"no-quorum-panic,attr,omitempty" json:"no-quorum-panic,attr,omitempty"`
	UpdateOrigin *string `xml:"update-origin,attr,omitempty" json:"update-origin,attr,omitempty"`
	UpdateClient *string `xml:"update-client,attr,omitempty" json:"update-client,attr,omitempty"`
	UpdateUser *string `xml:"update-user,attr,omitempty" json:"update-user,attr,omitempty"`
	ExecutionDate *string `xml:"execution-date,attr,omitempty" json:"execution-date,attr,omitempty"`
	Configuration Configuration `xml:"configuration" json:"configuration"`
	Status Status `xml:"status" json:"status"`
}

type cibAnyHolder struct {
	XMLName xml.Name
	XML     string `xml:",innerxml"`
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

type CibRuleExpression struct {
	XMLName xml.Name `xml:"expression" json:"-"`
	idMixin
	Attribute string `xml:"attribute,attr" json:"attribute"`
	Operation string `xml:"operation,attr" json:"operation"`
	Value *string `xml:"value,attr,omitempty" json:"value,omitempty"`
	Type *string `xml:"type,attr,omitempty" json:"type,omitempty"`
}

type dateCommonMixin struct {
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

type CibDateDuration struct {
	XMLName xml.Name `xml:"duration" json:"-"`
	dateCommonMixin
}

type CibDateSpec struct {
	XMLName xml.Name `xml:"date_spec" json:"-"`
	dateCommonMixin
}

type CibDateExpression struct {
	XMLName xml.Name `xml:"date_expression" json:"-"`
	idMixin
	Operation *string `xml:"operation,attr,omitempty" json:"operation,omitempty"`
	Start *string `xml:"start,attr,omitempty" json:"start,omitempty"`
	End *string `xml:"end,attr,omitempty" json:"end,omitempty"`
	Duration *CibDateDuration `xml:"duration,omitempty" json:"duration,omitempty"`
	DateSpec *CibDateSpec `xml:"date_spec,omitempty" json:"date_spec,omitempty"`
}

type CibRule struct {
	XMLName xml.Name `xml:"rule" json:"-"`
	idRefMixin
	scoreMixin
	ScoreAttribute *string `xml:"score-attribute,attr,omitempty" json:"score-attribute,omitempty"`
	BooleanOp *string `xml:"boolean-op,attr,omitempty" json:"boolean-op,omitempty"`
	Role *string `xml:"role,attr,omitempty" json:"role,omitempty" validate:"ValidAttributeRole"`
	Children []cibAnyHolder `xml:",any" json:"children,omitempty"`
}

type CibNVPair struct {
	XMLName xml.Name `xml:"nvpair" json:"-"`
	idRefMixin
	Name *string `xml:"name,attr,omitempty" json:"name,omitempty"`
	Value *string `xml:"value,attr,omitempty" json:"value,omitempty"`
}

type cibAttributeSetMixin struct {
	idRefMixin
	scoreMixin
	Children []cibAnyHolder `xml:",any" json:"children,omitempty"`
}

type CibInstanceAttributes struct {
	XMLName xml.Name `xml:"instance_attributes" json:"-"`
	cibAttributeSetMixin
}

type CibMetaAttributes struct {
	XMLName xml.Name `xml:"meta_attributes" json:"-"`
	cibAttributeSetMixin
}

type CibUtilization struct {
	XMLName xml.Name `xml:"utilization" json:"-"`
	cibAttributeSetMixin
}

type CibOp struct {
	XMLName xml.Name `xml:"op" json:"-"`
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
	Attributes []CibInstanceAttributes `xml:"instance_attributes" json:"instance_attributes,omitempty"`
	Meta []CibMetaAttributes `xml:"meta_attributes" json:"meta_attributes,omitempty"`
}

type CibOpSet struct {
	XMLName xml.Name `xml:"operations" json:"-"`
	idRefMixin
	Operations []CibOp `xml:"op" json:"operations,omitempty"`
}

type cibResourceMixin struct {
	idMixin
	descMixin
	Attributes []CibInstanceAttributes `xml:"instance_attributes" json:"instance_attributes,omitempty"`
	Meta []CibMetaAttributes `xml:"meta_attributes" json:"meta_attributes,omitempty"`
}


type CibTemplate struct {
	XMLName xml.Name `xml:"template" json:"-"`
	cibResourceMixin
	Type string `xml:"type,attr" json:"type"`
	Class string `xml:"class,attr" json:"class" validate:"ValidResourceClass"`
	Provider *string `xml:"provider,attr,omitempty" json:"provider,omitempty"`
	Utilization []CibUtilization `xml:"utilization" json:"utilization,omitempty"`
	Operations []CibOpSet `xml:"operations" json:"operations,omitempty"`
}


type CibPrimitive struct {
	XMLName xml.Name `xml:"primitive" json:"-"`
	cibResourceMixin
	Type *string `xml:"type,attr,omitempty" json:"type,omitempty"`
	Class *string `xml:"class,attr,omitempty" json:"class,omitempty" validate:"ValidResourceClass"`
	Provider *string `xml:"provider,attr,omitempty" json:"provider,omitempty"`
	Template *string `xml:"template,attr,omitempty" json:"template,omitempty"`
	Utilization []CibUtilization `xml:"utilization" json:"utilization,omitempty"`
	Operations []CibOpSet `xml:"operations" json:"operations,omitempty"`
}


type CibGroup struct {
	XMLName xml.Name `xml:"group" json:"-"`
	cibResourceMixin
	Children []CibPrimitive `xml:"primitive" json:"children"`
}

type CibClone struct {
	XMLName xml.Name `xml:"clone" json:"-"`
	cibResourceMixin
	Child cibAnyHolder `xml:",any" json:"child"`
}

type CibMaster struct {
	XMLName xml.Name `xml:"master" json:"-"`
	cibResourceMixin
	Child []cibAnyHolder `xml:",any" json:"child"`
}

type cibConstraintMixin struct {
	idMixin
}

type CibResourceRef struct {
	XMLName xml.Name `xml:"resource_ref" json:"-"`
	idMixin
}

type CibResourceSet struct {
	XMLName xml.Name `xml:"resource_set" json:"-"`
	idRefMixin
	Sequential *bool `xml:"sequential,attr,omitempty" json:"sequential,omitempty"`
	RequireAll *bool `xml:"require-all,attr,omitempty" json:"require-all,omitempty"`
	Ordering *string `xml:"ordering,attr,omitempty" json:"ordering,omitempty" validate:"ValidConstraintOrdering"`
	Action *string `xml:"action,attr,omitempty" json:"action,omitempty" validate:"ValidAttributeAction"`
	Role *string `xml:"role,attr,omitempty" json:"role,omitempty" validate:"ValidAttributeRole"`
	scoreMixin
	Resources []CibResourceRef `xml:"resource_ref" json:"resources,omitempty"`
}

type CibLocation struct {
	XMLName xml.Name `xml:"rsc_location" json:"-"`
	cibConstraintMixin
	Rsc *string `xml:"rsc,attr,omitempty" json:"rsc,omitempty"`
	RscPattern *string `xml:"rsc-pattern,attr,omitempty" json:"rsc-pattern,omitempty"`
	Role *string `xml:"role,attr,omitempty" json:"role,omitempty"`
	scoreMixin
	Node *string `xml:"node,attr,omitempty" json:"node,omitempty"`
	ResourceSets []CibResourceSet `xml:"resource_set" json:"resource-sets,omitempty"`
	ResourceDiscovery *string `xml:"resource-discovery,attr,omitempty" json:"resource-discovery,omitempty" validate:"ValidAttributeDiscovery"`
	Rules []CibRule `xml:"rule" json:"rules,omitempty"`
}

type CibColocation struct {
	XMLName xml.Name `xml:"rsc_colocation" json:"-"`
	cibConstraintMixin
	scoreMixin
	ResourceSets []CibResourceSet `xml:"resource_set" json:"resource-sets,omitempty"`
	Rsc *string `xml:"rsc,attr,omitempty" json:"rsc,omitempty"`
	WithRsc *string `xml:"with-rsc,attr,omitempty" json:"with-rsc,omitempty"`
	NodeAttribute *string `xml:"node-attribute,attr,omitempty" json:"node-attribute,omitempty"`
	RscRole *string `xml:"rsc-role,attr,omitempty" json:"rsc-role,omitempty"`
	WithRscRole *string `xml:"with-rsc-role,attr,omitempty" json:"with-rsc-role,omitempty"`
}

type CibOrder struct {
	XMLName xml.Name `xml:"rsc_order" json:"-"`
	cibConstraintMixin
	Symmetrical *bool `xml:"symmetrical,attr,omitempty" json:"symmetrical,omitempty"`
	RequireAll *bool `xml:"require-all,attr,omitempty" json:"require-all,omitempty"`
	scoreMixin
	Kind *string `xml:"kind,attr,omitempty" json:"kind,omitempty"`
	ResourceSets []CibResourceSet `xml:"resource_set" json:"resource-sets,omitempty"`
	First *string `xml:"first,attr,omitempty" json:"first,omitempty"`
	Then *string `xml:"then,attr,omitempty" json:"then,omitempty"`
	FirstAction *string `xml:"first-action,attr,omitempty" json:"first-action,omitempty" validate:"ValidAttributeAction"`
	ThenAction *string `xml:"then-action,attr,omitempty" json:"then-action,omitempty" validate:"ValidAttributeAction"`
}

type CibTicket struct {
	XMLName xml.Name `xml:"rsc_ticket" json:"-"`
	cibConstraintMixin
	ResourceSets []CibResourceSet `xml:"resource_set" json:"resource-sets,omitempty"`
	Rsc *string `xml:"rsc,attr,omitempty" json:"rsc,omitempty"`
	RscRole *string `xml:"rsc-role,attr,omitempty" json:"rsc-role,omitempty" validate:"ValidAttributeRole"`
	Ticket string `xml:"ticket,attr" json:"ticket"`
	LossPolicy *string `xml:"loss-policy,attr,omitempty" json:"loss-policy,omitempty" validate:"ValidTicketLossPolicy"`
}

type CibNode struct {
	XMLName xml.Name `xml:"node" json:"-"`
	idMixin
	descMixin
	scoreMixin
	Uname string `xml:"uname,attr" json:"uname"`
	Type *string `xml:"type,attr,omitempty" json:"type,omitempty" validate:"ValidNodeType"`
	Attributes []CibInstanceAttributes `xml:"instance_attributes" json:"instance_attributes,omitempty"`
	Utilization []CibUtilization `xml:"utilization" json:"utilization,omitempty"`
}

type CibFencingLevel struct {
	XMLName xml.Name `xml:"fencing-level" json:"-"`
	idMixin
	Target *string `xml:"target,attr,omitempty" json:"target,omitempty"`
	TargetPattern *string `xml:"target-pattern,attr,omitempty" json:"target-pattern,omitempty"`
	TargetAttribute *string `xml:"target-attribute,attr,omitempty" json:"target-attribute,omitempty"`
	TargetValue *string `xml:"target-value,attr,omitempty" json:"target-value,omitempty"`
	Index int `xml:"index,attr" json:"index"`
	Devices string `xml:"devices,attr" json:"devices"`
}

type CibRole struct {
	XMLName xml.Name `xml:"role" json:"-"`
	idMixin
}

type CibAclTarget struct {
	XMLName xml.Name `xml:"acl_target" json:"-"`
	idMixin
	Roles []CibRole `xml:"role" json:"roles,omitempty"`
}

type CibAclPermission struct {
	XMLName xml.Name `xml:"acl_permission" json:"-"`
	idMixin
	descMixin
	Kind string `xml:"kind,attr" json:"kind" validate:"ValidPermissionKind"`
	Xpath *string `xml:"xpath,attr,omitempty" json:"xpath,omitempty"`
	Reference *string `xml:"reference,attr,omitempty" json:"reference,omitempty"`
	ObjectType *string `xml:"object-type,attr,omitempty" json:"object-type,omitempty"`
	Attribute *string `xml:"attribute,attr,omitempty" json:"attribute,omitempty"`
}

type CibAclRole struct {
	XMLName xml.Name `xml:"acl_role" json:"-"`
	idMixin
	descMixin
	Permissions []CibAclPermission `xml:"acl_permission" json:"permissions,omitempty"`
}

type CibObjRef struct {
	XMLName xml.Name `xml:"obj_ref" json:"-"`
	idMixin
}

type CibTag struct {
	XMLName xml.Name `xml:"tag" json:"-"`
	idMixin
	References []CibObjRef `xml:"obj_ref" json:"references,omitempty"`
}

// can NOT be marshalled
type CibRecipient struct {
	XMLName xml.Name `xml:"recipient" json:"-"`
	idMixin
	descMixin
	Value string `xml:"value,attr" json:"value"`
	Meta []CibMetaAttributes `xml:"meta_attributes" json:"meta_attributes,omitempty"`
	Attributes []CibInstanceAttributes `xml:"instance_attributes" json:"instance_attributes,omitempty"`
}

// can NOT be marshalled
type CibAlert struct {
	XMLName xml.Name `xml:"alert" json:"-"`
	idMixin
	descMixin
	Path string `xml:"path,attr" json:"path"`
	Meta []CibMetaAttributes `xml:"meta_attributes" json:"meta_attributes,omitempty"`
	Attributes []CibInstanceAttributes `xml:"instance_attributes" json:"instance_attributes,omitempty"`
	Recipients []CibRecipient `xml:"recipient" json:"recipients,omitempty"`
}

type CibClusterPropertySet struct {
	XMLName xml.Name `xml:"cluster_property_set" json:"-"`
	cibAttributeSetMixin
}


// /cib/configuration
// needs custom parse + encode postprocess to recreate missing ids
type Configuration struct {
	CrmConfig []CibClusterPropertySet `xml:"crm_config>cluster_property_set" json:"crm_config,omitempty"`
	RscDefaults []CibMetaAttributes `xml:"rsc_defaults>meta_attributes" json:"rsc_defaults,omitempty"`
	OpDefaults []CibMetaAttributes `xml:"op_defaults>meta_attributes" json:"op_defaults,omitempty"`
	Nodes []CibNode `xml:"nodes>node" json:"nodes,omitempty"`
	Primitives []CibPrimitive `xml:"resources>primitive" json:"primitives,omitempty"`
	Groups []CibGroup `xml:"resources>group" json:"groups,omitempty"`
	Clones []CibClone `xml:"resources>clone" json:"clones,omitempty"`
	Masters []CibMaster `xml:"resources>master" json:"masters,omitempty"`
	Locations []CibLocation `xml:"constraints>rsc_location" json:"locations,omitempty"`
	Colocations []CibColocation `xml:"constraints>rsc_colocation" json:"colocations,omitempty"`
	Orders []CibOrder `xml:"constraints>rsc_order" json:"orders,omitempty"`
	Tickets []CibTicket `xml:"constraints>rsc_ticket" json:"tickets,omitempty"`
	Fencing []CibFencingLevel `xml:"fencing-topology>fencing-level" json:"fencing,omitempty"`
	AclTargets []CibAclTarget `xml:"acls>acl_target" json:"acl-targets,omitempty"`
	AclRoles []CibAclRole `xml:"acls>acl_role" json:"acl-roles,omitempty"`
	Tags []CibTag `xml:"tags>tag" json:"tags,omitempty"`
	Alerts []CibAlert `xml:"alerts>alert" json:"alerts,omitempty"`
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
	NodeState []NodeState `xml:"node_state" json:"node-state,omitempty"`
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

func (cib *Cib) ToJson() ([]byte, error) {
	return json.Marshal(cib)
}

func (status *Status) ToJson() ([]byte, error) {
	return json.Marshal(status)
}

func (configuration *Configuration) ToJson() ([]byte, error) {
	return json.Marshal(configuration)
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
// TODO: Generate object tree from RNG schema
func (cib *Cib) decodeCibObjects(xmldata []byte) error {
	return xml.Unmarshal(xmldata, &cib)
}
