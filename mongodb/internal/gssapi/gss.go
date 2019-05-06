// +build gssapi
// +build linux darwin

package gssapi

/*
#cgo linux CFLAGS: -DGOOS_linux
#cgo linux LDFLAGS: -lgssapi_krb5 -lkrb5
#cgo darwin CFLAGS: -DGOOS_darwin
#cgo darwin LDFLAGS: -framework GSS
#include "gss_wrapper.h"
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// NewServer creates a new SaslServer.
func NewServer(serviceName, hostname string, constrainedDelegation bool) *SaslServer {
	spn := fmt.Sprintf("%s@%s", serviceName, hostname)
	return &SaslServer{
		servicePrincipalName:  spn,
		constrainedDelegation: constrainedDelegation,
	}
}

type saslState int

// saslState constants
const (
	Start saslState = iota
	ContextComplete
	SupportComplete
	Done
)

// SaslServer hosts the server-side piece of a GSSAPI SASL mechanism conversation.
type SaslServer struct {
	servicePrincipalName string

	gss   C.mongosql_gssapi_server_state
	state saslState

	constrainedDelegation bool
}

// Close closes the SASL Server.
func (ss *SaslServer) Close() {
	C.mongosql_gssapi_server_destroy(&ss.gss)
}

// Start initializes the SASL Server.
func (ss *SaslServer) Start() error {
	cusername := C.CString(ss.servicePrincipalName)
	defer C.free(unsafe.Pointer(cusername))

	constrainedDelegationInt := C.int(0)
	if ss.constrainedDelegation {
		constrainedDelegationInt = C.int(1)
	}

	status := C.mongosql_gssapi_server_init(&ss.gss, cusername, constrainedDelegationInt)
	if status != C.GSSAPI_OK {
		return ss.getError("unable to initialize server")
	}

	return nil
}

// Next uses the challenge provided to continue with authentication and update the state of the server.
func (ss *SaslServer) Next(challenge []byte) ([]byte, error) {

	var buf unsafe.Pointer
	var bufLen C.size_t
	var outBuf unsafe.Pointer
	var outBufLen C.size_t

	switch ss.state {
	case Start:
		if len(challenge) > 0 {
			buf = unsafe.Pointer(&challenge[0])
			bufLen = C.size_t(len(challenge))
		}

		status := C.mongosql_gssapi_server_negotiate(&ss.gss, buf, bufLen, &outBuf, &outBufLen)
		switch status {
		case C.GSSAPI_OK:
			if !ss.constrainedDelegation && ss.gss.has_delegated_client_cred == 0 {
				return nil, fmt.Errorf("client did not provide a delegated credential")
			}
			ss.state = ContextComplete
		case C.GSSAPI_CONTINUE:
		default:
			return nil, ss.getError("unable to negotiate with client")
		}
	case ContextComplete:
		qop := []byte{1, 0, 0, 0} // quality of protection we support, which is none
		buf = unsafe.Pointer(&qop[0])
		bufLen = C.size_t(len(qop))
		status := C.mongosql_gssapi_server_wrap_msg(&ss.gss, buf, bufLen, &outBuf, &outBufLen)
		if status != C.GSSAPI_OK {
			return nil, ss.getError("unable to wrap security support")
		}

		ss.state = SupportComplete
	case SupportComplete:
		buf = unsafe.Pointer(&challenge[0])
		bufLen = C.size_t(len(challenge))
		status := C.mongosql_gssapi_server_unwrap_msg(&ss.gss, buf, bufLen, &outBuf, &outBufLen)
		if status != C.GSSAPI_OK {
			return nil, ss.getError("unable to unwrap authz")
		}

		ss.state = Done
	default:
		return nil, fmt.Errorf("Invalid state in SaslServer")
	}

	if outBuf != nil {
		defer C.free(outBuf)
	}

	return C.GoBytes(outBuf, C.int(outBufLen)), nil
}

// Completed indicates whether the server is in the "Done" state,
// ie. the authentication process is complete.
func (ss *SaslServer) Completed() bool {
	return ss.state == Done
}

func (ss *SaslServer) getError(prefix string) error {
	return getError(prefix, ss.gss.maj_stat, ss.gss.min_stat)
}

// NewClient creates a new SaslClient.
func NewClient(hostname string, server *SaslServer, serviceName string,
	canonicalizeHostName bool, serviceRealm string) *SaslClient {

	servicePrincipalName := fmt.Sprintf("%s@%s", serviceName, hostname)

	// ignore canonicalizeHostName and serviceRealm as they aren't supported by GSS

	return &SaslClient{
		servicePrincipalName: servicePrincipalName,
		serverGss:            server.gss,
	}
}

// SaslClient implements the client-side portion of a GSSAPI SASL Mechanism conversation.
type SaslClient struct {
	servicePrincipalName string
	serverGss            C.mongosql_gssapi_server_state

	gss   C.mongosql_gssapi_client_state
	state saslState
}

// Close closes the SASL Client.
func (sc *SaslClient) Close() {
	C.mongosql_gssapi_client_destroy(&sc.gss)
}

// Start initializes the SASL Client.
func (sc *SaslClient) Start() (string, []byte, error) {
	const mechName = "GSSAPI"

	cservicePrincipalName := C.CString(sc.servicePrincipalName)
	defer C.free(unsafe.Pointer(cservicePrincipalName))
	status := C.mongosql_gssapi_client_init(&sc.gss, &sc.serverGss, cservicePrincipalName)

	if status != C.GSSAPI_OK {
		return mechName, nil, sc.getError("unable to initialize client")
	}

	return mechName, nil, nil
}

// Next uses the challenge provided to continue with authentication and update the state of the client.
func (sc *SaslClient) Next(challenge []byte) ([]byte, error) {

	var buf unsafe.Pointer
	var bufLen C.size_t
	var outBuf unsafe.Pointer
	var outBufLen C.size_t

	switch sc.state {
	case Start:
		if len(challenge) > 0 {
			buf = unsafe.Pointer(&challenge[0])
			bufLen = C.size_t(len(challenge))
		}

		status := C.mongosql_gssapi_client_negotiate(&sc.gss, buf, bufLen, &outBuf, &outBufLen)
		switch status {
		case C.GSSAPI_OK:
			sc.state = ContextComplete
		case C.GSSAPI_CONTINUE:
		default:
			return nil, sc.getError("unable to negotiate with server")
		}
	case ContextComplete:
		var cusername *C.char
		status := C.mongosql_gssapi_client_username(&sc.gss, &cusername)
		if status != C.GSSAPI_OK {
			return nil, sc.getError("unable to acquire username")
		}
		defer C.free(unsafe.Pointer(cusername))
		username := C.GoString(cusername)

		msg := []byte{1, 0, 0, 0} // first bytes are the quality of protection we would like
		msg = append(msg, []byte(username)...)
		buf = unsafe.Pointer(&msg[0])
		bufLen = C.size_t(len(msg))
		status = C.mongosql_gssapi_client_wrap_msg(&sc.gss, buf, bufLen, &outBuf, &outBufLen)
		if status != C.GSSAPI_OK {
			return nil, sc.getError("unable to wrap authz")
		}

		sc.state = Done
	default:
		return nil, fmt.Errorf("Invalid state in SaslClient")
	}

	if outBuf != nil {
		defer C.free(outBuf)
	}

	return C.GoBytes(outBuf, C.int(outBufLen)), nil
}

// Completed indicates whether the client is in the "Done" state,
// ie. the authentication process is complete.
func (sc *SaslClient) Completed() bool {
	return sc.state == Done
}

func (sc *SaslClient) getError(prefix string) error {
	return getError(prefix, sc.gss.maj_stat, sc.gss.min_stat)
}

func getError(prefix string, maj, min C.OM_uint32) error {
	var desc *C.char

	status := C.mongosql_gssapi_error_desc(maj, min, &desc)
	if status != C.GSSAPI_OK {
		if desc != nil {
			C.free(unsafe.Pointer(desc))
		}

		return fmt.Errorf("%s: (%v, %v)", prefix, maj, min)
	}
	defer C.free(unsafe.Pointer(desc))

	return fmt.Errorf("%s: %v(%v,%v)", prefix, C.GoString(desc), int32(maj), int32(min))
}
