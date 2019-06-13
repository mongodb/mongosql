//+build gssapi,windows

#ifndef SSPI_WRAPPER_H
#define SSPI_WRAPPER_H

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
} sspi_client_state_old;

int sspi_init_old();

int sspi_client_init_old(
    sspi_client_state_old *client,
    char* username,
    char* password
);

int sspi_client_username_old(
    sspi_client_state_old *client,
    char** username
);

int sspi_client_negotiate_old(
    sspi_client_state_old *client,
    char* spn,
    PVOID input,
    ULONG input_length,
    PVOID* output,
    ULONG* output_length
);

int sspi_client_wrap_msg_old(
    sspi_client_state_old *client,
    PVOID input,
    ULONG input_length,
    PVOID* output,
    ULONG* output_length 
);

int sspi_client_destroy_old(
    sspi_client_state_old *client
);

#endif