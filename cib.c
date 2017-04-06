/*
  Copyright (C) 2017 Kristoffer Gronlund <kgronlund@suse.com>
  See LICENSE for license.
*/
#include "_cgo_export.h"

#define F_CIB_UPDATE_RESULT "cib_update_result"

int go_cib_signon(cib_t* cib, const char* name, enum cib_conn_type type) {
    return cib->cmds->signon(cib, name, type);
}

int go_cib_signoff(cib_t* cib) {
    return cib->cmds->signoff(cib);
}

int go_cib_query(cib_t * cib, const char *section, xmlNode ** output_data, int call_options) {
    return cib->cmds->query(cib, section, output_data, call_options);
}

static void go_cib_destroy_cb(gpointer user_data) {
	extern void destroyNotifyCallback();
	destroyNotifyCallback();
}

static cib_t *s_cib = NULL;
static xmlNode *s_current_cib = NULL;

static void go_cib_notify_cb(const char *event, xmlNode * msg) {
	int rc;
	rc = pcmk_ok;

	xmlNode *diff = get_message_xml(msg, F_CIB_UPDATE_RESULT);

	if (s_current_cib == NULL) {
		s_cib->cmds->query(s_cib, NULL, &s_current_cib, cib_scope_local | cib_sync_call);
	} else {
		rc = xml_apply_patchset(s_current_cib, diff, TRUE);
	}

	extern void diffNotifyCallback(xmlNode*);
	diffNotifyCallback(s_current_cib);
}


int go_cib_register_notify_callbacks(cib_t * cib) {
	int rc;

	s_cib = cib;
	s_current_cib = NULL;

	rc = cib->cmds->set_connection_dnotify(cib, go_cib_destroy_cb);
	if (rc != pcmk_ok) {
		return rc;
	}
	rc = cib->cmds->del_notify_callback(cib, T_CIB_DIFF_NOTIFY, go_cib_notify_cb);
	if (rc != pcmk_ok) {
		return rc;
	}
	rc = cib->cmds->add_notify_callback(cib, T_CIB_DIFF_NOTIFY, go_cib_notify_cb);
	if (rc != pcmk_ok) {
		return rc;
	}
	return pcmk_ok;
}

static gboolean idle_callback(gpointer user_data) {
	extern void goMainloopSched();
	goMainloopSched();
}

void go_add_idle_scheduler(GMainLoop* loop) {
	g_idle_add(&idle_callback, loop);
}
