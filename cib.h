#ifndef GOPACEMAKER_H_
#define GOPACEMAKER_H_

#include <crm/cib.h>
#include <crm/services.h>
#include <crm/common/util.h>
#include <crm/common/xml.h>
#include <crm/common/mainloop.h>

int go_cib_signon(cib_t* cib, const char* name, enum cib_conn_type type);
int go_cib_signoff(cib_t* cib);
int go_cib_query(cib_t * cib, const char *section, xmlNode ** output_data, int call_options);
int go_cib_register_notify_callbacks(cib_t * cib);
void go_add_idle_scheduler(GMainLoop* loop);

#endif/*GOPACEMAKER_H_*/
