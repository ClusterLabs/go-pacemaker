package pacemaker

import (
	"unsafe"
	"fmt"
)

/*
#cgo pkg-config: pacemaker pacemaker-cib libqb glib-2.0
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

type CibError struct {
	msg string
}

func (e *CibError) Error() string {
	return e.msg
}

func formatErrorRc(rc int) *CibError {
	return &CibError{fmt.Sprintf("%d: %s %s", rc, C.GoString(C.pcmk_errorname(C.int(rc))), C.GoString(C.pcmk_strerror(C.int(rc))))}
}

type CibConnection int

const (
	Query CibConnection = C.cib_query
	Command CibConnection = C.cib_command
)


type Cib struct {
	cCib *C.cib_t
}

type CibVersion struct {
	admin_epoch int32
	epoch int32
	num_updates int32
}

func (ver *CibVersion) String() string {
	return fmt.Sprintf("%d:%d:%d", ver.admin_epoch, ver.epoch, ver.num_updates)
}

func NewCib() *Cib {
	var cib Cib
	cib.cCib = C.cib_new()
	return &cib
}

func (cib *Cib) SignOn(connection CibConnection) *CibError {
	if cib.cCib.state == C.cib_connected_query || cib.cCib.state == C.cib_connected_command {
		return nil
	}

	rc := C.go_cib_signon(cib.cCib, C.crm_system_name, (uint32)(connection))
	if rc != C.pcmk_ok {
		return formatErrorRc((int)(rc))
	}
	return nil
}

func (cib *Cib) SignOff() *CibError {
	rc := C.go_cib_signoff(cib.cCib)
	if rc != C.pcmk_ok {
		return formatErrorRc((int)(rc))
	}
	return nil
}

func (cib *Cib) Delete() {
	C.cib_delete(cib.cCib)
}

func (cib *Cib) queryImpl(xpath string, nochildren bool) (*C.xmlNode, *CibError) {
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
		rc = C.go_cib_query(cib.cCib, C.CString(xpath), (**C.xmlNode)(unsafe.Pointer(&root)), opts)
	} else {
		rc = C.go_cib_query(cib.cCib, nil, (**C.xmlNode)(unsafe.Pointer(&root)), opts)
	}
	if rc != C.pcmk_ok {
		return nil, formatErrorRc((int)(rc))
	}
	return root, nil
}


func (cib *Cib) Version() (*CibVersion, *CibError) {
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


func (cib *Cib) Query() (string, *CibError) {
	var root *C.xmlNode
	root, err := cib.queryImpl("", false)
	if err != nil {
		return "", err
	}
	defer C.free_xml(root)

	buffer := C.dump_xml_unformatted(root)
	defer C.free(unsafe.Pointer(buffer))

	return C.GoString(buffer), nil
}

func (cib *Cib) QueryNoChildren() (string, *CibError) {
	var root *C.xmlNode
	root, err := cib.queryImpl("", true)
	if err != nil {
		return "", err
	}
	defer C.free_xml(root)

	buffer := C.dump_xml_unformatted(root)
	defer C.free(unsafe.Pointer(buffer))

	return C.GoString(buffer), nil
}


func (cib *Cib) QueryXPath(xpath string) (string, *CibError) {
	var root *C.xmlNode
	root, err := cib.queryImpl(xpath, false)
	if err != nil {
		return "", err
	}
	defer C.free_xml(root)

	buffer := C.dump_xml_unformatted(root)
	defer C.free(unsafe.Pointer(buffer))

	return C.GoString(buffer), nil
}

func (cib *Cib) QueryXPathNoChildren(xpath string) (string, *CibError) {
	var root *C.xmlNode
	root, err := cib.queryImpl(xpath, true)
	if err != nil {
		return "", err
	}
	defer C.free_xml(root)

	buffer := C.dump_xml_unformatted(root)
	defer C.free(unsafe.Pointer(buffer))

	return C.GoString(buffer), nil
}

func init() {
	C.crm_log_init(C.CString("go-pacemaker"), C.LOG_CRIT, 0, 0, 0, nil, 1)
}
