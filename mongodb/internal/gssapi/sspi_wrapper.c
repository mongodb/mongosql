//+build gssapi,windows

#include <stdio.h>
#include "sspi_wrapper.h"

static HINSTANCE sspi_secur32_dll = NULL;
static PSecurityFunctionTable sspi_functions = NULL;
static const LPSTR SSPI_PACKAGE_NAME = "kerberos";

int mongosql_sspi_wrap_msg(
    SECURITY_STATUS* status,
    CtxtHandle* ctx,
    PVOID input,
    ULONG input_length,
    PVOID* output,
    ULONG* output_length
)
{
    SecPkgContext_Sizes sizes;
    
    *status = sspi_functions->QueryContextAttributes(
        ctx,                // phContext
        SECPKG_ATTR_SIZES,  // ulAttribute
        &sizes              // pBuffer
    );
    if (*status != SEC_E_OK) {
        return SSPI_ERROR;
    }

    char *msg = malloc((sizes.cbSecurityTrailer + input_length + sizes.cbBlockSize) * sizeof(char));
    memcpy(&msg[sizes.cbSecurityTrailer], input, input_length);

    SecBuffer wrap_bufs[3];
    SecBufferDesc wrap_buf_desc;
    wrap_buf_desc.cBuffers = 3;
    wrap_buf_desc.pBuffers = wrap_bufs;
    wrap_buf_desc.ulVersion = SECBUFFER_VERSION;

    wrap_bufs[0].cbBuffer = sizes.cbSecurityTrailer;
    wrap_bufs[0].BufferType = SECBUFFER_TOKEN;
    wrap_bufs[0].pvBuffer = msg;

    wrap_bufs[1].cbBuffer = input_length;
    wrap_bufs[1].BufferType = SECBUFFER_DATA;
    wrap_bufs[1].pvBuffer = msg + sizes.cbSecurityTrailer;

    wrap_bufs[2].cbBuffer = sizes.cbBlockSize;
    wrap_bufs[2].BufferType = SECBUFFER_PADDING;
    wrap_bufs[2].pvBuffer = msg + sizes.cbSecurityTrailer + input_length;

    *status = sspi_functions->EncryptMessage(
        ctx,                    // phContext
        SECQOP_WRAP_NO_ENCRYPT, // fQOP
        &wrap_buf_desc,         // pMessage
        0                       // MessageSeqNo
    );
    if (*status != SEC_E_OK) {
        free(msg);
        return SSPI_ERROR;
    }

    *output_length = wrap_bufs[0].cbBuffer + wrap_bufs[1].cbBuffer + wrap_bufs[2].cbBuffer;
    *output = malloc(*output_length);

    memcpy(*output, wrap_bufs[0].pvBuffer, wrap_bufs[0].cbBuffer);
    memcpy(*output + wrap_bufs[0].cbBuffer, wrap_bufs[1].pvBuffer, wrap_bufs[1].cbBuffer);
    memcpy(*output + wrap_bufs[0].cbBuffer + wrap_bufs[1].cbBuffer, wrap_bufs[2].pvBuffer, wrap_bufs[2].cbBuffer);

    free(msg);

    return SSPI_OK;
}

int mongosql_sspi_init(
)
{
    sspi_secur32_dll = LoadLibrary("secur32.dll");
    if (!sspi_secur32_dll) {
        return GetLastError();
    }

    INIT_SECURITY_INTERFACE init_security_interface = (INIT_SECURITY_INTERFACE)GetProcAddress(sspi_secur32_dll, SECURITY_ENTRYPOINT);
    if (!init_security_interface) {
        return -1;
    }

    sspi_functions = (*init_security_interface)();
    if (!sspi_functions) {
        return -2;
    }

    return SSPI_OK;
}

int mongosql_sspi_client_init(
    mongosql_sspi_client_state *client,
    mongosql_sspi_server_state *server,
    char* username
)
{
    client->status = sspi_functions->ImpersonateSecurityContext(&server->ctx);
    if (client->status != SEC_E_OK) {
        return SSPI_ERROR;
    }

    TimeStamp timestamp;
    client->status = sspi_functions->AcquireCredentialsHandle(
        username,               // pszPrincipal
        SSPI_PACKAGE_NAME,      // pszPackage
        SECPKG_CRED_OUTBOUND,   // fCredentialUse
        NULL,                   // pvLogonID
        NULL,                   // pAuthData
        NULL,                   // pGetKeyFn
        NULL,                   // pvGetKeyArgument
        &client->cred,          // phCredential
        &timestamp              // ptsExpiry
    );

    if (client->status != SEC_E_OK) {
        return SSPI_ERROR;
    }

    client->status = sspi_functions->RevertSecurityContext(&server->ctx);
    if (client->status != SEC_E_OK) {
        return SSPI_ERROR;
    }

    return SSPI_OK;
}

int mongosql_sspi_client_negotiate(
    mongosql_sspi_client_state *client,
    char* spn,
    PVOID input,
    ULONG input_length,
    PVOID* output,
    ULONG* output_length
)
{
    SecBufferDesc inbuf;
    SecBuffer in_bufs[1];
    SecBufferDesc outbuf;
    SecBuffer out_bufs[1];

    if (client->has_ctx > 0) {
        inbuf.ulVersion = SECBUFFER_VERSION;
        inbuf.cBuffers = 1;
        inbuf.pBuffers = in_bufs;
        in_bufs[0].pvBuffer = input;
        in_bufs[0].cbBuffer = input_length;
        in_bufs[0].BufferType = SECBUFFER_TOKEN;
    }

    outbuf.ulVersion = SECBUFFER_VERSION;
    outbuf.cBuffers = 1;
    outbuf.pBuffers = out_bufs;
    out_bufs[0].pvBuffer = NULL;
    out_bufs[0].cbBuffer = 0;
    out_bufs[0].BufferType = SECBUFFER_TOKEN;

    ULONG context_attr = 0;

    client->status = sspi_functions->InitializeSecurityContext(
        &client->cred,                                  // phCredential
        client->has_ctx > 0 ? &client->ctx : NULL,      // phContext
        (LPSTR) spn,                                    // pszTargetName
        ISC_REQ_ALLOCATE_MEMORY | ISC_REQ_MUTUAL_AUTH,  // fContextReq
        0,                                              // Reserved1
        SECURITY_NETWORK_DREP,                          // TargetDataRep
        client->has_ctx > 0 ? &inbuf : NULL,            // pInput
        0,                                              // Reserved2
        &client->ctx,                                   // phNewContext
        &outbuf,                                        // pOutput
        &context_attr,                                  // pfContextAttr
        NULL                                            // ptsExpiry
    );

    if (client->status != SEC_E_OK && client->status != SEC_I_CONTINUE_NEEDED) {
        return SSPI_ERROR;
    }

    client->has_ctx = 1;

    *output = malloc(out_bufs[0].cbBuffer);
    *output_length = out_bufs[0].cbBuffer;
    memcpy(*output, out_bufs[0].pvBuffer, *output_length);
    SECURITY_STATUS ss = sspi_functions->FreeContextBuffer(out_bufs[0].pvBuffer);
    if (ss != SEC_E_OK) {
        free(*output);
        client->status = ss;
        return SSPI_ERROR;
    }

    if (client->status == SEC_I_CONTINUE_NEEDED) {
        return SSPI_CONTINUE;
    }

    return SSPI_OK;
}

int mongosql_sspi_client_wrap_msg(
    mongosql_sspi_client_state *client,
    PVOID input,
    ULONG input_length,
    PVOID* output,
    ULONG* output_length 
)
{
    return mongosql_sspi_wrap_msg(
        &client->status,
        &client->ctx,
        input,
        input_length,
        output,
        output_length
    );
}

int mongosql_sspi_client_destroy(
    mongosql_sspi_client_state *client,
    mongosql_sspi_server_state *server
)
{
    SECURITY_STATUS ss;
    int result = SSPI_OK;
    if (client->has_ctx > 0) {
        ss = sspi_functions->DeleteSecurityContext(&client->ctx);
        if (ss != SEC_E_OK) {
            result = SSPI_ERROR;
        }
    }

    ss = sspi_functions->FreeCredentialsHandle(&client->cred);
    if (ss != SEC_E_OK) {
        result = SSPI_ERROR;
    }

    return result;
}

int mongosql_sspi_server_init(
    mongosql_sspi_server_state *server,
    char* username
)
{
    TimeStamp timestamp;
    
    server->status = sspi_functions->AcquireCredentialsHandle(
        username,               // pszPrincipal
        SSPI_PACKAGE_NAME,      // pszPackage
        SECPKG_CRED_BOTH,       // fCredentialUse
        NULL,                   // pvLogonID
        NULL,                   // pAuthData
        NULL,                   // pGetKeyFn
        NULL,                   // pvGetKeyArgument
        &server->cred,          // phCredential
        &timestamp              // ptsExpiry
    );

    if (server->status != SEC_E_OK) {
        return SSPI_ERROR;
    }
    
    return SSPI_OK;
}

int mongosql_sspi_server_negotiate(
    mongosql_sspi_server_state *server,
    PVOID input,
    ULONG input_length,
    PVOID* output,
    ULONG* output_length 
)
{
    SecBufferDesc inbuf;
    SecBuffer in_bufs[1];
    SecBufferDesc outbuf;
    SecBuffer out_bufs[1];

    inbuf.ulVersion = SECBUFFER_VERSION;
    inbuf.cBuffers = 1;
    inbuf.pBuffers = in_bufs;
    in_bufs[0].pvBuffer = input;
    in_bufs[0].cbBuffer = input_length;
    in_bufs[0].BufferType = SECBUFFER_TOKEN;

    outbuf.ulVersion = SECBUFFER_VERSION;
    outbuf.cBuffers = 1;
    outbuf.pBuffers = out_bufs;
    out_bufs[0].pvBuffer = NULL;
    out_bufs[0].cbBuffer = 0;
    out_bufs[0].BufferType = SECBUFFER_TOKEN;

    ULONG context_attr = 0;
    
    server->status = sspi_functions->AcceptSecurityContext(
        &server->cred,                                                      // phCredential
        server->has_ctx > 0 ? &server->ctx : NULL,                          // phContext
        &inbuf,                                                             // pInput
        ASC_REQ_ALLOCATE_MEMORY | ASC_REQ_MUTUAL_AUTH | ASC_REQ_DELEGATE,   // fContextReq
        SECURITY_NETWORK_DREP,                                              // TargetDataRep
        &server->ctx,                                                       // phNewContext
        &outbuf,                                                            // pOutput
        &context_attr,                                                      // pfContextAttr
        NULL                                                                // ptsTimeStamp
    );

    if (server->status != SEC_E_OK && server->status != SEC_I_CONTINUE_NEEDED) {
        return SSPI_ERROR;
    }

    server->has_ctx = 1;

    *output = malloc(out_bufs[0].cbBuffer);
    *output_length = out_bufs[0].cbBuffer;
    memcpy(*output, out_bufs[0].pvBuffer, *output_length);
    SECURITY_STATUS ss = sspi_functions->FreeContextBuffer(out_bufs[0].pvBuffer);
    if (ss != SEC_E_OK) {
        free(*output);
        server->status = ss;
        return SSPI_ERROR;
    }

    if (server->status == SEC_I_CONTINUE_NEEDED) {
        return SSPI_CONTINUE;
    }

    return SSPI_OK;
}

int mongosql_sspi_server_wrap_msg(
    mongosql_sspi_server_state *server,
    PVOID input,
    ULONG input_length,
    PVOID* output,
    ULONG* output_length 
)
{
    return mongosql_sspi_wrap_msg(
        &server->status,
        &server->ctx,
        input,
        input_length,
        output,
        output_length
    );
}

int mongosql_sspi_server_unwrap_msg(
    mongosql_sspi_server_state *server,
    PVOID input,
    ULONG input_length,
    PVOID* output,
    ULONG* output_length 
)
{
    SecBufferDesc buf;
    SecBuffer bufs[2];
    buf.cBuffers = 2;
    buf.pBuffers = bufs;
    buf.ulVersion = SECBUFFER_VERSION;

    bufs[0].cbBuffer = input_length;
    bufs[0].BufferType = SECBUFFER_STREAM;
    bufs[0].pvBuffer = input;

    bufs[1].cbBuffer = 0;
    bufs[1].BufferType = SECBUFFER_DATA;
    bufs[1].pvBuffer = NULL;

    ULONG qop = 0;

    server->status = sspi_functions->DecryptMessage(
        &server->ctx,   // phContext
        &buf,           // pMessage
        0,              // MessageSeqNo
        &qop            // pfQOP
    );

    if (server->status != SEC_E_OK) {
        return SSPI_ERROR;
    }

    *output = malloc(bufs[1].cbBuffer);
    *output_length = bufs[1].cbBuffer;
    memcpy(*output, bufs[1].pvBuffer, *output_length);
    
    return SSPI_OK;
}

int mongosql_sspi_server_destroy(
    mongosql_sspi_server_state *server
)
{
    SECURITY_STATUS ss;
    int result = SSPI_OK;
    if (server->has_ctx > 0) {
        ss = sspi_functions->DeleteSecurityContext(&server->ctx);
        if (ss != SEC_E_OK) {
            result = SSPI_ERROR;
        }
    }

    ss = sspi_functions->FreeCredentialsHandle(&server->cred);
    if (ss != SEC_E_OK) {
        result = SSPI_ERROR;
    }

    return result;
}
