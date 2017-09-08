# GSSAPI

## General Description

[GSSAPI][gssapi] is a meta-protocol that is mainly used to wrap the kerberos protocol. MongoDB and, by extension, MongoSQL, utilize SASL which is also a meta-protocol. In effect, the SASL meta-protocol is wrapping the GSSAPI meta-protocol which is wrapping the Kerberos protocol. This is the [GSSAPI SASL Mechanism][gssapi-sasl].

Overall, SASL works by (a) beginning a conversation and (b) continuing that conversatiom until the server and client agree it is completed.

GSSAPI works in largely the same way. A conversation is begun by initializing a security context. Sometimes this takes more than one round trip. Once that has been completed, then everything is secure. However, SASL imposes an additional roundtrip on the GSSAPI protocol which carries with it information about the username of the client and message confidentiality/integrity settings. MongoDB does not support any of these, so we send the username and bytes indicating not confidentiality or integrity. These messages are sent using Encrypt/Wrap and Decrypt/Unwrap depending on the implementation. The wrapper classes, discussed below, distill all this into a common API such that the different implementations look remarkably similar.

### Client <-> MongoSQL

In this part of the authentication stages, we are acting as the server and utilize the server portions of the GSSAPI protocol. This conversation is wrapped by the type `gssapi.SaslServer`.

Over the custom auth protocol, we only require a single conversation with the client. This is because we are implementing Delegated Authentication. In effect, the client gives MongoSQL permission to use their credentials to talk with MongoDB on their behalf. 

### MongoSQL <-> MongoDB

In this part of the authentictation stages, we are acting as the client and utilize the client portions of the GSSAPI protocol. This is wrapped by the type `gssapi.SaslClient`.

This conversation is implemented in the mongo-go-driver by calling auth.ConductSaslConversation. At this point, we let the driver handle negotiation with the server. Because the credentials were delegated, we can conduct a multiple conversation with MongoDB even though only one conversation happened with the client.

## Implementations

The separate implementations are wrappers and, as such, some things have been abstracted out and hidden. For instance, for each implementation, the return codes have been boiled down to the 3 necessary to make progress. XXX_OK, XXX_CONTINUE, and XXX_ERROR. Native error codes are provided in the client and server state. This is done to make the cgo calls easier. The alternative would be (depending on implementation), yet another call back into C to tell if this was in fact an error.

### Linux / MacOS

The gss api is implemented in the files prefixed by `gss`.

* [gss_wrapper.c](gss_wrapper.c) and [gss_wrapper.h](gss_wrapper.h) handle wrapping the actual C api. It binds to the standard gssapi headers that are implemented by both MIT and heimdal (and on MacOS, the GSS framework which should be Heimdal). Ultimately, the goal of these wrapper files is twofold.
    1. make it easier to call into C. There are a number of types that are easier wrangled in C than in Go.
    2. Store state in a C struct. This lessens the burden on go to hold onto certain pointers and structs that are difficult to translate. 

* [gss.go](gss.go) is the go wrapper around the gss_wrapper files. It handles all the type conversions and memory wrangling and implements the higher level protocol of encoding and decoding messages passed between either the client and mongosqld or mongosqld and the server.

## References

* [GSSAPI][gssapi]
* [GSSAPI SASL Mechanism][gssapi-sasl]

[gssapi]: https://tools.ietf.org/html/rfc2743 "GSSAPI"
[gssapi-sasl]: https://tools.ietf.org/html/rfc4752  "GSSAPI SASL Mechanism"
