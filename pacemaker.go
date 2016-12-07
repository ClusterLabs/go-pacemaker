// The pacemaker package provides an API for reading the Pacemaker cluster configuration (CIB).
package pacemaker

import (
	"unsafe"
	"fmt"
	"encoding/xml"
	"encoding/json"
	"strings"
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

type CibOpenConfig struct {
	connection CibConnection
	file string
	shadow string
	server string
	user string
	passwd string
	port int
	encrypted bool
}

func ForQuery(config *CibOpenConfig) {
	config.connection = Query
}

func ForCommand(config *CibOpenConfig) {
	config.connection = Command
}

func FromFile(file string) func(*CibOpenConfig) {
	return func(config *CibOpenConfig) {
		config.file = file
	}
}

func FromShadow(shadow string) func(*CibOpenConfig) {
	return func(config *CibOpenConfig) {
		config.shadow = shadow
	}
}

func FromRemote(server, user, passwd string, port int, encrypted bool) func (*CibOpenConfig) {
	return func(config *CibOpenConfig) {
		config.server = server
		config.user = user
		config.passwd = passwd
		config.port = port
		config.encrypted = encrypted
	}
}

type Element struct {
	Type string
	Id string
	Attr map[string]string
	Elements []*Element
}

// Root entity representing the CIB. Can be
// populated with CIB data if the Decode
// method is used.
type Cib struct {
	cCib *C.cib_t
	Attr map[string]string
	Config *Element
	Status Status `xml:"status" json:"status"`
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
	TransitionKey string `xml:"transition-key,attr" json:"transition-key,omitempty"`
	TransitionMagic string `xml:"transition-magic,attr" json:"transition-magic,omitempty"`
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

type TransitionMagic struct {
	Op *ResourceStateOp
	Uuid string
	TransitionId int
	ActionId int
	OpStatus int
	OpRc int
	TargetRc int
}

func (op *ResourceStateOp) DecodeTransitionMagic() TransitionMagic {
	magic := C.CString(op.TransitionMagic)
	defer C.free(unsafe.Pointer(magic))
	var uuid *C.char
	var transition_id C.int
	var action_id C.int
	var op_status C.int
	var op_rc C.int
	var target_rc C.int
	C.decode_transition_magic(magic, &uuid, &transition_id, &action_id, &op_status, &op_rc, &target_rc)
	return TransitionMagic{
		op,
		C.GoString(uuid),
		(int)(transition_id),
		(int)(action_id),
		(int)(op_status),
		(int)(op_rc),
		(int)(target_rc),
	}
}

func (status *Status) ResourceStatus(id string) string {
	state := "stopped"
	for _, node := range status.NodeState {
		for _, rsc := range node.Resources {
			if rsc.Id != id {
				continue
			}
			for _, op := range rsc.Ops {
				if op.Rc == 0 {
				}
			}
		}
	}
	return state
}

func OpenCib(options ...func (*CibOpenConfig)) (*Cib, error) {
	var cib Cib
	config := CibOpenConfig{}
	for _, opt := range options {
		opt(&config)
	}
	if config.connection != Query && config.connection != Command {
		config.connection = Query
	}
	if config.file != "" {
		s := C.CString(config.file)
		defer C.free(unsafe.Pointer(s))
		cib.cCib = C.cib_file_new(s)
	} else if config.shadow != "" {
		s := C.CString(config.shadow)
		defer C.free(unsafe.Pointer(s))
		cib.cCib = C.cib_shadow_new(s)
	} else if config.server != "" {
		s := C.CString(config.server)
		u := C.CString(config.user)
		p := C.CString(config.passwd)
		defer C.free(unsafe.Pointer(s))
		defer C.free(unsafe.Pointer(u))
		defer C.free(unsafe.Pointer(p))
		var e C.int = 0
		if config.encrypted {
			e = 1
		}
		cib.cCib = C.cib_remote_new(s, u, p, (C.int)(config.port), (C.gboolean)(e))
	} else {
		cib.cCib = C.cib_new()
	}

	rc := C.go_cib_signon(cib.cCib, C.crm_system_name, (uint32)(config.connection))
	if rc != C.pcmk_ok {
		return nil, formatErrorRc((int)(rc))
	}

	return &cib, nil
}

func GetShadowFile(name string) string {
	s := C.CString(name)
	defer C.free(unsafe.Pointer(s))
	return C.GoString(C.get_shadow_file(s))
}

func (cib *Cib) Close() error {
	rc := C.go_cib_signoff(cib.cCib)
	if rc != C.pcmk_ok {
		return formatErrorRc((int)(rc))
	}
	C.cib_delete(cib.cCib)
	cib.cCib = nil
	return nil
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


func IsTrue(bstr string) bool {
	sl := strings.ToLower(bstr)
	return sl == "true" || sl == "on" || sl == "yes" || sl == "y" || sl == "1"
}


func (cib *Cib) decodeCibObjects(xmldata []byte) error {
	decoder := xml.NewDecoder(bytes.NewReader(xmldata))
	for {
		t, err := decoder.Token()
		if t == nil {
			return nil
		} else if err != nil {
			return err
		}
		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "cib" {
				cib.Attr = make(map[string]string)
				for _, attr := range se.Attr {
					cib.Attr[attr.Name.Local] = attr.Value
				}
			} else if  se.Name.Local == "status" {
				decoder.DecodeElement(&cib.Status, &se)
			} else if se.Name.Local == "configuration" {
			}
		case xml.EndElement:
		}
	}
}
