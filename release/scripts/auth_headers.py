"""
This file is a modified version of the sample linked below. It creates the authorization
request headers for visiting a given s3 host and URI with a HEAD request (instead
of a GET request). It will not execute the HEAD request, but instead return
the headers to the caller. It is hardcoded to have no query string.
"""

# Copyright 2010-2018 Amazon.com, Inc. or its affiliates. All Rights Reserved.
#
# This file is licensed under the Apache License, Version 2.0 (the "License").
# You may not use this file except in compliance with the License. A copy of the
# License is located at
#
# http://aws.amazon.com/apache2.0/
#
# This file is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS
# OF ANY KIND, either express or implied. See the License for the specific
# language governing permissions and limitations under the License.
#
# ABOUT THIS PYTHON SAMPLE: This sample is part of the AWS General Reference
# Signing AWS API Requests top available at
# https://docs.aws.amazon.com/general/latest/gr/sigv4-signed-request-examples.html
#

# AWS Version 4 signing example

# See: http://docs.aws.amazon.com/general/latest/gr/sigv4_signing.html
# This version makes a HEAD request and passes the signature
# in the Authorization header.
import sys
import os
import datetime
import hashlib
import hmac

# ************* REQUEST VALUES *************
METHOD = 'HEAD'
SERVICE = 's3'
REGION = 'us-east-1'

def construct_headers_for_head(host, canonical_uri):
    """
    Constructs the AWS authorization headers for a HEAD request to the given host and URI.
    """

    # Key derivation functions. See:
    # http://docs.aws.amazon.com/general/latest/gr/signature-v4-examples.html#signature-v4-examples-python
    def sign(key, msg):
        """
        Sign message using key.
        """
        return hmac.new(key, msg.encode('utf-8'), hashlib.sha256).digest()

    def get_signature_key(key, datestamp, region_name, service_name):
        """
        Derive the signing key from arguments.
        """
        k_date = sign(('AWS4' + key).encode('utf-8'), datestamp)
        k_region = sign(k_date, region_name)
        k_service = sign(k_region, service_name)
        k_signing = sign(k_service, 'aws4_request')
        return k_signing

    # Read AWS access key from env. variables or configuration file. Best practice is NOT
    # to embed credentials in code.
    access_key = os.environ.get('AWS_ACCESS_KEY_ID')
    if access_key is None:
        print "Can't find 'AWS_ACCESS_KEY_ID' in environment variables"
        sys.exit(1)

    secret_key = os.environ.get('AWS_SECRET_ACCESS_KEY')
    if secret_key is None:
        print "Can't find 'AWS_SECRET_ACCESS_KEY' in environment variables"
        sys.exit(1)

    # Create a date for headers and the credential string
    now = datetime.datetime.utcnow()
    amzdate = now.strftime('%Y%m%dT%H%M%SZ')
    datestamp = now.strftime('%Y%m%d') # Date w/o time, used in credential scope


    # ************* TASK 1: CREATE A CANONICAL REQUEST *************
    # http://docs.aws.amazon.com/general/latest/gr/sigv4-create-canonical-request.html

    # Step 1 is to define the verb (GET, POST, etc.)--already done.

    # Step 2: Create canonical URI--the part of the URI from domain to query
    # string (use '/' if no path)
    # This is taken in as an argument -- already done.

    # Step 3: Create the canonical query string. In this example (a HEAD request),
    # request parameters are in the query string. Query string values must
    # be URL-encoded (space=%20). The parameters must be sorted by name.
    # In this case, no query string is used.
    canonical_querystring = ""

    # Step 4: Create payload hash (hash of the request body content).
    # In this case, the payload is an empty string ("").
    payload_hash = hashlib.sha256(('').encode('utf-8')).hexdigest()


    # Step 5: Create the canonical headers and signed headers. Header names
    # must be trimmed and lowercase, and sorted in code point order from
    # low to high. Note that there is a trailing \n.
    canonical_headers = ("host:%s\nx-amz-content-sha256:%s\nx-amz-date:%s\n"
                         % (host, payload_hash, amzdate))

    # Step 6: Create the list of signed headers. This lists the headers
    # in the canonical_headers list, delimited with ";" and in alpha order.
    # Note: The request can include any headers; canonical_headers and
    # signed_headers lists those that you want to be included in the
    # hash of the request. "Host", "x-amz-date", and "x-amz-content-sha256"
    # are  required.
    # x-amz-content-sha256 is required when using signature version 4 to
    # authenticate request; this header provides a hash of the request payload.
    # https://docs.aws.amazon.com/AmazonS3/latest/API/RESTCommonRequestHeaders.html
    signed_headers = 'host;x-amz-content-sha256;x-amz-date' #

    # Step 7: Combine elements to create canonical request
    canonical_request = "\n".join([METHOD, canonical_uri, canonical_querystring,
                                   canonical_headers, signed_headers, payload_hash])

    # ************* TASK 2: CREATE THE STRING TO SIGN*************
    # Match the algorithm to the hashing algorithm you use, either SHA-1 or
    # SHA-256 (recommended)
    algorithm = 'AWS4-HMAC-SHA256'
    credential_scope = '/'.join([datestamp, REGION, SERVICE, 'aws4_request'])
    string_to_sign = '\n'.join([algorithm, amzdate, credential_scope,
                                hashlib.sha256(canonical_request.encode('utf-8')).hexdigest()])

    # ************* TASK 3: CALCULATE THE SIGNATURE *************
    # Create the signing key using the function defined above.
    signing_key = get_signature_key(secret_key, datestamp, REGION, SERVICE)

    # Sign the string_to_sign using the signing_key
    signature = hmac.new(signing_key, (string_to_sign).encode('utf-8'), hashlib.sha256).hexdigest()

    # ************* TASK 4: ADD SIGNING INFORMATION TO THE REQUEST *************
    # The signing information can be either in a query string value or in
    # a header named Authorization. This code shows how to use a header.
    # Create authorization header and add to request headers.
    authorization_header = ("%s Credential=%s/%s, SignedHeaders=%s, Signature=%s"
                            % (algorithm, access_key, credential_scope, signed_headers, signature))

    # The request can include any headers, but MUST include "host", "x-amz-date",
    # "x-amz-content-sha256", and (for this scenario) "Authorization". "host",
    # "x-amz-date", and "x-amz-content-sha256" must be included in the
    # canonical_headers and signed_headers, as noted earlier. Order here is not
    # significant.
    # Python note: The 'host' header is added automatically by the Python 'requests' library.

    headers = {'x-amz-date':amzdate,
               'Authorization':authorization_header,
               'x-amz-content-sha256':payload_hash}

    # We return the headers used in a HEAD request by the caller.
    # This allows the caller to be responsible for handling the errors and
    # information returned by requests.head(url, headers).
    return headers
