## mongodb/
This package is the primary location of code that is used to communicate with MongoDB
servers.

This README covers basics about the "high level" and "low level" Go driver APIs, why we
use the "low level" API, and the difference between "admin" and "user" sessions.

### Go Driver APIs
Our main use case for the Go driver is to be able to run commands against a mongo server.
It offers 2 APIs to do this.

#### "High Level"
(We do not use this API, so feel free to skip this section if you are not interested!)
The "high level" API that most users of the driver consume provides the `mongo.Client`
type. A `mongo.Client` can be created with a connection string (mongodb uri) or a
programmatic set of options that represents the same info as a connection string. The
connection string (or options) provided to a `mongo.Client` at creation time may contain
authentication info (username, password, mechanism, auth properties, auth source). That
auth info is turned into an appropriate "handshaker" function that is used each time the
`mongo.Client` communicates with the server. The `mongo.Client` has a few useful methods,
but the most relevant one is the `Database(name string)` method which returns a
`mongo.Database`. The `mongo.Database` type has the `RunCommand()` method and the
`Collection()` method, which are used for running database and collection level commands.
If authentication is enabled in the client (i.e., auth info was provided when the client
was created), the "handshaker" is used whenever a command is run. (Technically, that
_may_ not be the case--the `mongo.Client` has a pool of connections which may or may not
be authenticated already. That's an unimportant detail for this API, though, that is
abstracted away. To a consumer of `mongo.Client`, authentication can be set and will be
handled appropriately for all commands.)

#### "Low Level"
The "low level" API that _we_ consume provides several interfaces for communicating with
mongo servers. The main 3 are `driver.Deployment`, `driver.Server`, and
`driver.Connection`. There are also various `operation.*` types for running commands. It
also has a network library and a few other useful packages, but those are not relevant
for this README. To run a command using the low level API, a user can use the appropriate
`operation` type. (_Technically_, all commands are run using the `driver.Operation` type,
but there are useful wrappers such as `operation.Insert` and `operation.Aggregate` that
abstract that away). For a `driver.Operation` to execute, it needs a `driver.Deployment`
against which the operation runs, as well as a database (if necessary), a collection (if
necessary), a write concern (if necessary and desired), a read preference (if necessary
and desired), etc. When the operation is executed, it selects a `driver.Server` from the
`driver.Deployment`, then gets a `driver.Connection` from that server, and finally writes
a wire message on that connection and reads a wire message back as a response.

### `sqlproxy`'s Use Case
The first important thing to point out: whenever `sqlproxy` needs to communicate with
MongoDB it does that via the `mongodb.Session` type in this package. At `mongosqld` (or
`mongodrdl`) startup, a `mongodb.SessionProvider` is created using configuration options
provided by a user. That `SessionProvider` has three methods for providing `Session`s:
`Session()`, `AuthenticatedAdminSessionPrimary()`, and `AuthenticatedAdminSession()`. All
three methods simply return a `Session`, but the first one returns a "user" session while
the latter two return "admin" sessions. We'll get to the _why_ of this next idea in the
following paragraph: the key difference between "user" and "admin" `Session`s is the
`deployment` field's value. For a "user" session, the deployment is a
`driver.SingleServerDeployment`; for an "admin" session, the deployment is a
`topology.Topology`. "User" sessions handle `mongosqld`-user queries and (most) other
commands. "Admin" sessions handle administrative tasks such as sampling, updating schema,
etc. 

So, why do we need to use different deployment types for "user" vs. "admin" sessions?
The reason is authentication. The "admin" tasks are performed using auth information
that is provided in the config at `mongosqld` (or `mongodrdl`) start time. Because that
auth info is given to the `SessionProvider`, a `topology.Topology` can be created with
a "handshaker" that uses that auth info. (This is the `adminConnToplogy` field in the
`SessionProvider`). For "user" sessions, which are created whenever a client connects to
`mongosqld`, the connection may auth directly from the client to the mongod. Examples of
this are SCRAM and kerberos. For this type of authentication, we do not store any auth
credentials, so our only chance to get authenticated `driver.Connection`s to mongo is
during the mysql handshake. (The `userConnTopology` field of the `SessionProvider` stores
a topology with all the same settings as the admin one except it _does not_ have the
handshaker since its auth is handled differently).

Finally, how does this work in the code? As alluded to above, when a `SessionProvider` is
created, two topologies are created: `adminConnTopology` and `userConnTopology`. They use
the same connection uri and settings; the only difference is `adminConnTopology` has a
handshaker set which uses the admin auth info in the `SessionProvider` config.
- When an "admin" session is requested from a `SessionProvider`, a `Session` is returned
that uses the `adminConnTopology` as its backing deployment. That way, when commands are
run, a server is selected from that deployment, then connections are retrieved from that
server and auth is handled for those connections using the handshaker.
- When a "user" session is requested from a `SessionProvider`, a `Session` is returned
that uses a `driver.SingleServerDeployment` as its backing deploymnet. The "single
server" is the `Session` itself! Our `Session` type maintains its _own_ pool of
authenticated `driver.Connection`s and we use those connections to run commands.
`Session` implements `driver.Server` by implementing a `Connection(...)
driver.Connection` method. When commands are run on this type of "user" session, the
server selected from that deployment is the `Session` itself, and the connections
provided by that `Session` are the authenticated ones stored in its pool. Where do
those authenticated connections come from? Before we create the `Session`, we select
a server from `userConnTopology` and use that server to get connections. Those are
then added to the `Session`'s pool and authenticated via the `(Session).Login()` method.

##### Final note:
While we _could_ (and previously _did_) just do everything the "user" way, there exist
some circumstances where this can cause errors. The specific motivating example that lead
to us adding the `adminConnToplogy` was that if an election occurred during sampling, the
connections stored in the admin session would become invalid and we had no way of getting
new, authenticated connections. The `driver.SingleServerDeployment` which stores one of
our `Session`s only has a few already-authenticated `driver.Connection`s pooled. If those
connections become invalid, we cannot get new ones. The `adminConnTopology` that we now
use to run admin operations selects servers from the whole topology, so when we say
"run this operation against the primary", it will run against the correct primary even if
an election happens!
