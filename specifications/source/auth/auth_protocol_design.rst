====================================================
Design: MongoDB MySQL Authentication Plugin Protocol
====================================================

**********
References
**********
- `MongoDB authentication specification  <https://github.com/mongodb/specifications/blob/master/source/auth/auth.rst>`_

- `MySQL authentication docs <https://dev.mysql.com/doc/internals/en/authentication-method.html>`_

- `C Authentication Plugin <https://github.com/mongodb/mongosql-auth-c>`_

- `Java Authentication Plugin <https://github.com/mongodb/mongosql-auth-java>`_

****************
Initial Exchange
****************
1.  S - Initial Handshake

    1.  Plugin-name - ``mongosql_auth``

    2.  Auth-data

        1.  Major/Minor Version

            + 2 Bytes

            + Actually amounts to the 2 bytes plus 19 NUL bytes due to MySQL wire protocol.

***************
SASL
***************
1.  C - Handshake Response

    1.  Plugin Data

        1. No data - 0 bytes

2.  S - First Auth More Data

    1.  On first reply

        1.  Mechanism name

            +  Nul-terminated string

        2.  Mechanism-specific information

            +   SCRAM-SHA-1

                1. Number of conversations

            +   PLAIN

                1. Number of conversations

3.  C - Continued

    1.  Repeat for number of authentication conversations being conducted

        1.  Status

            +   Byte

                1.  0 = continue

                2.  1 = done

                    + On error, terminate connection

        2.  Payload length

            +  4 byte integer

        3.  Payload

            + bytes
4.  S - Continued

    1.  Repeat for number of authentication conversations being conducted

        1.  Payload length

            + 4 byte integer

        2.  Payload

            + Bytes
