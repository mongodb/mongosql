# Authentication Mechanisms

Authentication Mechanisms currently supported by the mongodb-schema-manager are most of those
supported by the MongoDB Rust driver on which the mongodb-schema-manager is built. The only
exception at this point is the Workforce flow for OIDC. A more in-depth representation of supported
Authentication mechanisms can be seen
[here](https://www.mongodb.com/docs/drivers/rust/current/fundamentals/authentication/)

Briefly, MongoDB Authentication Mechanisms have 5 attributes that may, must, or must not be set
based on the chosen Mechanism. These are:

- mechanism: the name of the Authentication Mechanism itself.
- username: the username for the user connecting to the database, perhaps surprisingly, some forms
  of authentication actually do not use this.
- password: this is password for the user.
- source: this is the database against which Authentication should be performed. In many mechanisms,
  this will be $external which means that Authentication is actually performed outside of MongoDB.
  Many other mechanisms allow specific credential information to be set in specific databases,
  though in practice this is often just the `admin` database.
- mechanism\_properties: these are additional key values pairs used by some select Mechanisms.

These attributes are set in the uri passed for connection, a guide on how to set these attributes in
the uri can be seen
[here](https://www.mongodb.com/docs/manual/reference/connection-string-options/#authentication-options),
the entire documentation for uris can be found
[here](https://www.mongodb.com/docs/manual/reference/connection-string/)

Mechanisms:
- SCRAM-SHA-1: Allows for connecting with username and password protected using SCRAM-SHA-1
    - mechanism MUST be "SCRAM-SHA-1"
    - username MUST be specified and non-zero length.
    - password MUST be specified.
    - source MUST be specified. Defaults to the database name if supplied on the connection string or admin.
    - mechanism\_properties MUST NOT be specified.

- SCRAM-SHA-256: Allows for connecting with username and password protected using SCRAM-SHA-256
    - mechanism MUST be "SCRAM-SHA-256"
    - username MUST be specified and non-zero length.
    - password MUST be specified.
    - source MUST be specified. Defaults to the database name if supplied on the connection string or admin.
    - mechanism\_properties MUST NOT be specified.

- MONGODB-X509: Allows for connecting via an x509 certificate, see the
[MongoDB documentation](https://www.mongodb.com/docs/manual/core/security-x.509/) for more information.
    - mechanism MUST be "MONGODB-X509"
    - username SHOULD NOT be provided for MongoDB 3.4+ MUST be specified and non-zero length for MongoDB prior to 3.4
    - password MUST NOT be specified.
    - source MUST be "$external". Defaults to $external.
    - mechanism\_properties MUST NOT be specified.

- GSSAPI (kerberos)
    - Not currently supported

- MONGODB-AWS
    - mechanism MUST be "MONGODB-AWS"
    - username MAY be specified. The non-sensitive AWS access key.
    - password MAY be specified. The sensitive AWS secret key.
    - source MUST be "$external". Defaults to $external.
    - mechanism\_properties
        - AWS\_SESSION\_TOKEN allows the user to specify an AWS session token for authentication with temporary credentials.

- MONGODB-OIDC: Allows for connecting via [OpenID Connect](https://openid.net/developers/specs/)
  access tokens. Currently, only Workload authentication in the azure or gcp environments and
  Workforce authentication are supported
    - username MAY be specified. Its meaning varies depending on the OIDC provider integration used.
    - source MUST be "$external". Defaults to $external.
    - password MUST NOT be specified.
    - mechanism MUST be "MONGODB-OIDC"
    - mechanism\_properties
        - ENVIRONMENT allows the user to specify the name of a built-in OIDC application environment integration to use to obtain Workload credentials. The value MUST be one of \["azure", "gcp"\].
            - If not set, Workforce authentication is assumed.
        - TOKEN\_RESOURCE (currently required, if ENVIRONMENT is set) the URI of the target resource.

- PLAIN (LDAP)
    - mechanism MUST be "PLAIN"
    - username MUST be specified and non-zero length.
    - password MUST be specified.
    - source MUST be specified. Defaults to the database name if supplied on the connection string or $external.
    - mechanism\_properties MUST NOT be specified.

