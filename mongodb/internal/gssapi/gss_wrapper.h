// +build gssapi
// +build linux darwin

#ifndef MONGOSQL_GSS_WRAPPER_H
#define MONGOSQL_GSS_WRAPPER_H

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
} mongosql_gssapi_client_state;

typedef struct {
    gss_cred_id_t cred;
    gss_ctx_id_t ctx;

    int has_delegated_client_cred;
    gss_cred_id_t delegated_client_cred;

    OM_uint32 maj_stat;
    OM_uint32 min_stat;
} mongosql_gssapi_server_state;

int mongosql_gssapi_error_desc(
    OM_uint32 maj_stat, 
    OM_uint32 min_stat, 
    char **desc
);

int mongosql_gssapi_client_init(
    mongosql_gssapi_client_state *client,
    mongosql_gssapi_server_state *server,
    char* spn
);

int mongosql_gssapi_client_username(
    mongosql_gssapi_client_state *client,
    char** username
);

int mongosql_gssapi_client_negotiate(
    mongosql_gssapi_client_state *client,
    void* input,
    size_t input_length,
    void** output,
    size_t* output_length
);

int mongosql_gssapi_client_wrap_msg(
    mongosql_gssapi_client_state *client,
    void* input,
    size_t input_length,
    void** output,
    size_t* output_length 
);

int mongosql_gssapi_client_destroy(
    mongosql_gssapi_client_state *client
);

int mongosql_gssapi_server_init(
    mongosql_gssapi_server_state *server,
    char* username
);

int mongosql_gssapi_server_negotiate(
    mongosql_gssapi_server_state *server,
    void* input,
    size_t input_length,
    void** output,
    size_t* output_length 
);

int mongosql_gssapi_server_wrap_msg(
    mongosql_gssapi_server_state *server,
    void* input,
    size_t input_length,
    void** output,
    size_t* output_length 
);

int mongosql_gssapi_server_unwrap_msg(
    mongosql_gssapi_server_state *server,
    void* input,
    size_t input_length,
    void** output,
    size_t* output_length 
);

int mongosql_gssapi_server_destroy(
    mongosql_gssapi_server_state *server
);

#endif // MONGOSQL_GSS_WRAPPER_H
