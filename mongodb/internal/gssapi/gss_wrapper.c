// +build gssapi
// +build linux darwin

#include <string.h>
#include <stdio.h>
#include "gss_wrapper.h"

OM_uint32 mongosql_copy_and_release_buffer(
    OM_uint32* minor_status,
    gss_buffer_desc* buffer,
    void** output,
    size_t* output_length
)
{
    OM_uint32 major_status;
    *output = malloc(buffer->length);
    *output_length = buffer->length;

    if (*output) {
        memcpy(*output, buffer->value, buffer->length);
    }

    major_status = gss_release_buffer(
        minor_status,   // minor_status
        buffer         // buffer
    );

    if (GSS_ERROR(major_status) && *output) {
        free(*output);
    }

    return major_status;
}

OM_uint32 mongosql_gssapi_canonicalize_name(
    OM_uint32* minor_status, 
    char *input_name, 
    gss_OID input_name_type, 
    gss_name_t *output_name
)
{
    OM_uint32 major_status;
    gss_name_t imported_name = GSS_C_NO_NAME;
    gss_buffer_desc input_name_buffer = GSS_C_EMPTY_BUFFER;

    input_name_buffer.value = input_name;
    input_name_buffer.length = strlen(input_name);
    major_status = gss_import_name(
        minor_status,       // minor_status
        &input_name_buffer, // input_name_buffer
        input_name_type,    // input_name_type
        &imported_name      // output_name
    );

    if (GSS_ERROR(major_status)) {
        return major_status;
    }

    major_status = gss_canonicalize_name(
        minor_status,           // minor_status
        imported_name,          // input_name
        (gss_OID)gss_mech_krb5, // mech_type
        output_name             // output_name
    );

    if (imported_name != GSS_C_NO_NAME) {
        major_status = gss_release_name(
            minor_status,   // minor_status
            &imported_name  // name
        );
    }

    return major_status;
}

OM_uint32 mongosql_gssapi_display_name(
    OM_uint32 *minor_status,
    gss_cred_id_t cred,
    char** output_name
)
{
    OM_uint32 major_status;
    OM_uint32 ignore;
    gss_name_t name = GSS_C_NO_NAME;

    major_status = gss_inquire_cred(
        minor_status,   // minor_status
        cred,           // cred_handle
        &name,          // name
        NULL,           // lifetime
        NULL,           // cred_usage
        NULL            // mechanisms
    );

    if (GSS_ERROR(major_status)) {
        return major_status;
    }

    gss_buffer_desc name_buffer;
    major_status = gss_display_name(minor_status, name, &name_buffer, NULL);
    if (GSS_ERROR(major_status)) {

        // ignore result here because it would mask the error from display name
        gss_release_name(
            &ignore,    // minor_status
            &name       // name
        );

        return major_status;
    }

    major_status = gss_release_name(
        minor_status,   // minor_status
        &name);         // name

    if (GSS_ERROR(major_status)) {
        if (name_buffer.length) {
            // ignore result here because it would mask the error from release_name
            gss_release_buffer(
                &ignore,   // minor_status
                &name_buffer    // buffer
            );
        }
        
        return major_status;
    }

    if (name_buffer.length) {
        *output_name = malloc(name_buffer.length+1); 
        memcpy(*output_name, name_buffer.value, name_buffer.length+1);

        major_status = gss_release_buffer(
            minor_status,   // minor_status
            &name_buffer    // buffer
        );

        if (GSS_ERROR(major_status) && *output_name) {
            free(*output_name);
        }
    }

    return major_status;
}

OM_uint32 mongosql_gssapi_wrap_msg(
    OM_uint32 *minor_status,
    gss_ctx_id_t ctx,
    void* input,
    size_t input_length,
    void** output,
    size_t* output_length 
)
{
    OM_uint32 major_status;
    gss_buffer_desc input_buffer = GSS_C_EMPTY_BUFFER;
    gss_buffer_desc output_buffer = GSS_C_EMPTY_BUFFER;

    input_buffer.value = input;
    input_buffer.length = input_length;

    major_status = gss_wrap(
        minor_status,       // minor_status
        ctx,                // context_handle
        0,                  // conf_req_flag
        GSS_C_QOP_DEFAULT,  // qop_req
        &input_buffer,      // input_message_buffer
        NULL,               // conf_state
        &output_buffer      // output_message_buffer
    );

    if (GSS_ERROR(major_status)) {
        return major_status;
    }

    if (output_buffer.length) {
        major_status = mongosql_copy_and_release_buffer(
            minor_status,
            &output_buffer,
            output,
            output_length
        );
    }

    return major_status;
}

int mongosql_gssapi_error_desc(
    OM_uint32 maj_stat, 
    OM_uint32 min_stat, 
    char **desc
)
{
    OM_uint32 stat = maj_stat;
    int stat_type = GSS_C_GSS_CODE;
    if (min_stat != 0) {
        stat = min_stat;
        stat_type = GSS_C_MECH_CODE;
    }

    OM_uint32 local_maj_stat, local_min_stat;
    OM_uint32 msg_ctx = 0;
    gss_buffer_desc desc_buffer;

    // error codes are stored in a hierarchical fashion. This loop
    // will overwrite more general errors with more specific errors.
    do
    {
        if (*desc) {
            free(*desc);
        }
        
        local_maj_stat = gss_display_status(
            &local_min_stat,    // minor_status
            stat,               // status_value
            stat_type,          // status_type
            GSS_C_NO_OID,       // mech_type
            &msg_ctx,           // message_context
            &desc_buffer        // status_string
        );

        if (GSS_ERROR(local_maj_stat)) {
            return GSSAPI_ERROR;
        }

        *desc = malloc(desc_buffer.length+1);
        memcpy(*desc, desc_buffer.value, desc_buffer.length+1);

        local_maj_stat = gss_release_buffer(
            &local_min_stat,    // minor_status
            &desc_buffer        // buffer
        );

        if (GSS_ERROR(local_maj_stat)) {
            if (*desc) {
                free(*desc);
            }
            return GSSAPI_ERROR;
        }
    }
    while(msg_ctx != 0);

    return GSSAPI_OK;
}

int mongosql_gssapi_client_init(
    mongosql_gssapi_client_state *client,
    mongosql_gssapi_server_state *server,
    char* spn
)
{
    client->cred = server->delegated_client_cred;
    client->ctx = GSS_C_NO_CONTEXT;

    client->maj_stat = mongosql_gssapi_canonicalize_name(
        &client->min_stat, 
        spn, 
        GSS_C_NT_HOSTBASED_SERVICE, 
        &client->spn
    );

    if (GSS_ERROR(client->maj_stat)) {
        return GSSAPI_ERROR;
    }
    return GSSAPI_OK;
}

int mongosql_gssapi_client_username(
    mongosql_gssapi_client_state *client,
    char** username
)
{
    client->maj_stat = mongosql_gssapi_display_name(
        &client->min_stat, 
        client->cred, 
        username
    );

    if (GSS_ERROR(client->maj_stat)) {
        return GSSAPI_ERROR;
    }
    return GSSAPI_OK;
}

int mongosql_gssapi_client_negotiate(
    mongosql_gssapi_client_state *client,
    void* input,
    size_t input_length,
    void** output,
    size_t* output_length
)
{
    gss_buffer_desc input_buffer = GSS_C_EMPTY_BUFFER;
    gss_buffer_desc output_buffer = GSS_C_EMPTY_BUFFER;

    if (input) {
        input_buffer.value = input;
        input_buffer.length = input_length;
    }

    client->maj_stat = gss_init_sec_context(
        &client->min_stat,          // minor_status
        client->cred,               // initiator_cred_handle
        &client->ctx,               // context_handle
        client->spn,                // target_name
        GSS_C_NO_OID,               // mech_type
        GSS_C_MUTUAL_FLAG | GSS_C_INTEG_FLAG, // req_flags
        0,                          // time_req
        GSS_C_NO_CHANNEL_BINDINGS,  // input_chan_bindings
        &input_buffer,              // input_token
        NULL,                       // actual_mech_type
        &output_buffer,             // output_token
        NULL,                       // ret_flags
        NULL                        // time_rec
    );

    if (GSS_ERROR(client->maj_stat)) {
        return GSSAPI_ERROR;
    }

    if (output_buffer.length) {
        OM_uint32 major_status;
        OM_uint32 minor_status;
        major_status = mongosql_copy_and_release_buffer(
            &minor_status,
            &output_buffer,
            output,
            output_length
        );

        if (GSS_ERROR(major_status)) {
            client->maj_stat = major_status;
            client->min_stat = minor_status;
            return GSSAPI_ERROR;
        }
    }

    if (client->maj_stat == GSS_S_CONTINUE_NEEDED) {
        return GSSAPI_CONTINUE;
    }
    return GSSAPI_OK;
}

int mongosql_gssapi_client_wrap_msg(
    mongosql_gssapi_client_state *client,
    void* input,
    size_t input_length,
    void** output,
    size_t* output_length 
)
{
    client->maj_stat = mongosql_gssapi_wrap_msg(
        &client->min_stat, 
        client->ctx, 
        input, 
        input_length, 
        output, 
        output_length
    );

    if (GSS_ERROR(client->maj_stat)) {
        return GSSAPI_ERROR;
    }
    return GSSAPI_OK;
}

int mongosql_gssapi_client_destroy(
    mongosql_gssapi_client_state *client
)
{
    int result = GSSAPI_OK;
    OM_uint32 major_status;
    OM_uint32 minor_status;
    if (client->ctx != GSS_C_NO_CONTEXT) {
        major_status = gss_delete_sec_context(
            &minor_status,      // minor_status
            &client->ctx,       // context_handle
            GSS_C_NO_BUFFER     // output_token
        );

        if (GSS_ERROR(major_status)) {
            result = GSSAPI_ERROR;
        }
    }

    if (client->spn != GSS_C_NO_NAME) {
        major_status = gss_release_name(
            &minor_status,  // minor_status
            &client->spn    // name
        );
        if (GSS_ERROR(major_status)) {
            result = GSSAPI_ERROR;
        }
    }

    // NOTE: do not release client->cred because these are delegated credentials
    // held onto by mongosql_gssapi_server_state. They will be released when the server
    // is destroyed.

    return result;
}

int mongosql_gssapi_server_init(
    mongosql_gssapi_server_state *server,
    char* username
)
{
    server->cred = GSS_C_NO_CREDENTIAL;
    server->ctx = GSS_C_NO_CONTEXT;
    server->delegated_client_cred = GSS_C_NO_CREDENTIAL;

    gss_name_t spn = GSS_C_NO_NAME;
    server->maj_stat = mongosql_gssapi_canonicalize_name(
        &server->min_stat, 
        username, 
        GSS_C_NT_HOSTBASED_SERVICE, 
        &spn
    );

    if (GSS_ERROR(server->maj_stat)) {
        return GSSAPI_ERROR;
    }

    server->maj_stat = gss_acquire_cred(
        &server->min_stat,  // minor_status
        spn,                // desired_name
        GSS_C_INDEFINITE,   // time_req
        GSS_C_NO_OID_SET,   // desired_mech
        GSS_C_BOTH,         // cred_usage
        &server->cred,      // output_cred_handle
        NULL,               // actual_mechs
        NULL                // time_rec
    );

    if (spn != GSS_C_NO_NAME) {
        OM_uint32 major_status;
        OM_uint32 minor_status;
        major_status = gss_release_name(
            &minor_status,  // minor_status
            &spn            // name
        );

        // ignore this local error if gss_acquire_cred had an error
        if (GSS_ERROR(major_status) && !GSS_ERROR(server->maj_stat)) {
            server->maj_stat = major_status;
            server->min_stat = minor_status;
        }
    }

    if (GSS_ERROR(server->maj_stat)) {
        return GSSAPI_ERROR;
    }

    return GSSAPI_OK;
}

int mongosql_gssapi_server_negotiate(
    mongosql_gssapi_server_state *server,
    void* input,
    size_t input_length,
    void** output,
    size_t* output_length 
)
{
    gss_buffer_desc input_buffer = GSS_C_EMPTY_BUFFER;
    gss_buffer_desc output_buffer = GSS_C_EMPTY_BUFFER;
    OM_uint32 ret_flags;

    if (input) {
        input_buffer.value = input;
        input_buffer.length = input_length;
    }

    server->maj_stat = gss_accept_sec_context(
        &server->min_stat,              // minor_status
        &server->ctx,                   // context_handle
        server->cred,                   // acceptor_cred_handle
        &input_buffer,                  // input_token
        GSS_C_NO_CHANNEL_BINDINGS,      // input_chan_bindings
        NULL,                           // src_name
        NULL,                           // mech_type
        &output_buffer,                 // output_token
        &ret_flags,                     // ret_flags
        NULL,                           // time_rec
        &server->delegated_client_cred  // delegated_cred_handle
    );

    if (GSS_ERROR(server->maj_stat)) {
        return GSSAPI_ERROR;
    }

    if (output_buffer.length) {
        OM_uint32 major_status;
        OM_uint32 minor_status;
        major_status = mongosql_copy_and_release_buffer(
            &minor_status,
            &output_buffer,
            output,
            output_length
        );

        if (GSS_ERROR(major_status)) {
            server->maj_stat = major_status;
            server->min_stat = minor_status;
            return GSSAPI_ERROR;
        }
    }

    if (server->maj_stat == GSS_S_CONTINUE_NEEDED) {
        return GSSAPI_CONTINUE;
    }

    return GSSAPI_OK;
}

int mongosql_gssapi_server_wrap_msg(
    mongosql_gssapi_server_state *server,
    void* input,
    size_t input_length,
    void** output,
    size_t* output_length 
)
{
    server->maj_stat = mongosql_gssapi_wrap_msg(
        &server->min_stat, 
        server->ctx, 
        input, 
        input_length, 
        output, 
        output_length
    );

    if (GSS_ERROR(server->maj_stat)) {
        return GSSAPI_ERROR;
    }

    return GSSAPI_OK;
}

int mongosql_gssapi_server_unwrap_msg(
    mongosql_gssapi_server_state *server,
    void* input,
    size_t input_length,
    void** output,
    size_t* output_length 
)
{
    gss_buffer_desc input_buffer = GSS_C_EMPTY_BUFFER;
    gss_buffer_desc output_buffer = GSS_C_EMPTY_BUFFER;

    input_buffer.value = input;
    input_buffer.length = input_length;

    server->maj_stat = gss_unwrap(
        &server->min_stat,  // minor_status
        server->ctx,        // context_handle
        &input_buffer,      // input_message_buffer
        &output_buffer,     // output_message_buffer
        NULL,               // conf_state
        NULL                // qop_state
    );

    if (GSS_ERROR(server->maj_stat)) {
        return GSSAPI_ERROR;
    }

    if (output_buffer.length) {
        server->maj_stat = mongosql_copy_and_release_buffer(
            &server->min_stat,
            &output_buffer,
            output,
            output_length
        );

        if (GSS_ERROR(server->maj_stat)) {
            return GSSAPI_ERROR;
        }
    }

    return GSSAPI_OK;
}

int mongosql_gssapi_server_destroy(
    mongosql_gssapi_server_state *server
)
{
    OM_uint32 major_status;
    OM_uint32 minor_status;
    int result = GSSAPI_OK;
    if (server->ctx != GSS_C_NO_CONTEXT) {
        major_status = gss_delete_sec_context(
            &minor_status,  // minor_status
            &server->ctx,   // context_handle
            GSS_C_NO_BUFFER // output_token
        );

        if (GSS_ERROR(major_status)) {
            result = GSSAPI_ERROR;
        }
    }

    if (server->cred != GSS_C_NO_CREDENTIAL) {
        major_status = gss_release_cred(
            &minor_status,  // minor_status
            &server->cred   // cred_handle
        );

        if (GSS_ERROR(major_status)) {
            result = GSSAPI_ERROR;
        }
    }

    if (server->delegated_client_cred != GSS_C_NO_CREDENTIAL) {
        major_status = gss_release_cred(
            &minor_status,                  // minor_status
            &server->delegated_client_cred  // cred_handle
        );

        if (GSS_ERROR(major_status)) {
            result = GSSAPI_ERROR;
        }
    }

    return result;
}
