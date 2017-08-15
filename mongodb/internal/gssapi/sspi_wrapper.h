//+build gssapi,windows

#ifndef MONGOSQL_SSPI_WRAPPER_H
#define MONGOSQL_SSPI_WRAPPER_H

#define SECURITY_WIN32 1  /* Required for SSPI */

#include <windows.h>
#include <sspi.h>

#define SSPI_OK 0
#define SSPI_CONTINUE 1
#define SSPI_ERROR 2

typedef struct {
    CredHandle cred;
    CtxtHandle ctx;

    int has_ctx;

    SECURITY_STATUS status;
} mongosql_sspi_client_state;

typedef struct {
    CredHandle cred;
    CtxtHandle ctx;

    int has_ctx;

    SECURITY_STATUS status;
} mongosql_sspi_server_state;

int mongosql_sspi_init();

int mongosql_sspi_client_init(
    mongosql_sspi_client_state *client,
    mongosql_sspi_server_state *server,
    char* username
);

int mongosql_sspi_client_negotiate(
    mongosql_sspi_client_state *client,
    char* spn,
    PVOID input,
    ULONG input_length,
    PVOID* output,
    ULONG* output_length
);

int mongosql_sspi_client_wrap_msg(
    mongosql_sspi_client_state *client,
    PVOID input,
    ULONG input_length,
    PVOID* output,
    ULONG* output_length 
);

int mongosql_sspi_client_destroy(
    mongosql_sspi_client_state *client,
    mongosql_sspi_server_state *server
);

int mongosql_sspi_server_init(
    mongosql_sspi_server_state *server,
    char* username
);

int mongosql_sspi_server_username(
    mongosql_sspi_server_state *server,
    char** username
);

int mongosql_sspi_server_negotiate(
    mongosql_sspi_server_state *server,
    PVOID input,
    ULONG input_length,
    PVOID* output,
    ULONG* output_length 
);

int mongosql_sspi_server_revert(
    mongosql_sspi_server_state *server
);

int mongosql_sspi_server_wrap_msg(
    mongosql_sspi_server_state *server,
    PVOID input,
    ULONG input_length,
    PVOID* output,
    ULONG* output_length 
);

int mongosql_sspi_server_unwrap_msg(
    mongosql_sspi_server_state *server,
    PVOID input,
    ULONG input_length,
    PVOID* output,
    ULONG* output_length 
);

int mongosql_sspi_server_destroy(
    mongosql_sspi_server_state *server
);

#endif // MONGOSQL_SSPI_WRAPPER_H