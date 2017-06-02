/*
  Copyright (C) 2017 Kristoffer Gronlund <kgronlund@suse.com>
  See LICENSE for license.
*/
#include "_cgo_export.h"

#define F_CIB_UPDATE_RESULT "cib_update_result"

static cib_t *s_cib = NULL;

int go_cib_signon(cib_t* cib, const char* name, enum cib_conn_type type) {
	int rc;
	rc = cib->cmds->signon(cib, name, type);
	return rc;
}

int go_cib_signoff(cib_t* cib) {
	int rc;
	rc = cib->cmds->signoff(cib);
	return rc;
}

int go_cib_query(cib_t * cib, const char *section, xmlNode ** output_data, int call_options) {
	int rc;
	rc = cib->cmds->query(cib, section, output_data, call_options);
	return rc;
}

static void go_cib_destroy_cb(gpointer user_data) {
	extern void destroyNotifyCallback();
	destroyNotifyCallback();
}

static void go_cib_notify_cb(const char *event, xmlNode * msg) {
	int rc;
	rc = pcmk_ok;

	xmlNode *current_cib;
	xmlNode *diff = get_message_xml(msg, F_CIB_UPDATE_RESULT);

	s_cib->cmds->query(s_cib, NULL, &current_cib, cib_scope_local | cib_sync_call);

	extern void diffNotifyCallback(xmlNode*);
	diffNotifyCallback(current_cib);

	free_xml(current_cib);
}


int go_cib_register_notify_callbacks(cib_t * cib) {
	int rc;

	s_cib = cib;

	rc = cib->cmds->set_connection_dnotify(cib, go_cib_destroy_cb);
	if (rc != pcmk_ok) {
		goto exit;
	}
	rc = cib->cmds->del_notify_callback(cib, T_CIB_DIFF_NOTIFY, go_cib_notify_cb);
	if (rc != pcmk_ok) {
		goto exit;
	}
	rc = cib->cmds->add_notify_callback(cib, T_CIB_DIFF_NOTIFY, go_cib_notify_cb);
	if (rc != pcmk_ok) {
		goto exit;
	}
exit:
	return pcmk_ok;
}

static gboolean idle_callback(gpointer user_data) {
	extern void goMainloopSched();
	goMainloopSched();
}

void go_add_idle_scheduler(GMainLoop* loop) {
	g_idle_add(&idle_callback, loop);
}
