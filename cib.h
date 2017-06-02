#ifndef GOPACEMAKER_H_
#define GOPACEMAKER_H_

#include <crm/cib.h>
#include <crm/services.h>
#include <crm/common/util.h>
#include <crm/common/xml.h>
#include <crm/common/mainloop.h>

/*
  Flags returned by go_cib_register_notify_callbacks
  indicating which notifications were actually
  available to register (different connection types
  enable different sets of notifications)
 */
#define GO_CIB_NOTIFY_DESTROY 0x1
#define GO_CIB_NOTIFY_ADDREMOVE 0x2

int go_cib_signon(cib_t* cib, const char* name, enum cib_conn_type type);
int go_cib_signoff(cib_t* cib);
int go_cib_query(cib_t * cib, const char *section, xmlNode ** output_data, int call_options);
unsigned int go_cib_register_notify_callbacks(cib_t * cib);
void go_add_idle_scheduler(GMainLoop* loop);

#endif/*GOPACEMAKER_H_*/
