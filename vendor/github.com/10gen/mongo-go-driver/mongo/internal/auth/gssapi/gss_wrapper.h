//+build gssapi
//+build linux darwin
#ifndef GSS_WRAPPER_H
#define GSS_WRAPPER_H

#include <stdlib.h>
#ifdef GOOS_linux
#include <gssapi/gssapi.h>
#include <gssapi/gssapi_krb5.h>
#endif
#ifdef GOOS_darwin
#include <GSS/GSS.h>
#endif

#define GSSAPI_OK 0
#define GSSAPI_CONTINUE 1
#define GSSAPI_ERROR 2

typedef struct {
    gss_name_t spn;
    gss_cred_id_t cred;
    gss_ctx_id_t ctx;

    OM_uint32 maj_stat;
    OM_uint32 min_stat;
} gssapi_client_state_old;

int gssapi_error_desc_old(
    OM_uint32 maj_stat, 
    OM_uint32 min_stat, 
    char **desc
);

int gssapi_client_init_old(
    gssapi_client_state_old *client,
    char* spn,
    char* username,
    char* password
);

int gssapi_client_username_old(
    gssapi_client_state_old *client,
    char** username
);

int gssapi_client_negotiate_old(
    gssapi_client_state_old *client,
    void* input,
    size_t input_length,
    void** output,
    size_t* output_length
);

int gssapi_client_wrap_msg_old(
    gssapi_client_state_old *client,
    void* input,
    size_t input_length,
    void** output,
    size_t* output_length 
);

int gssapi_client_destroy_old(
    gssapi_client_state_old *client
);

#endif