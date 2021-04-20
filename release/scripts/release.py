"""
This script supports the release process for the BI Connector.
"""

import ast
import datetime
import getopt
import hashlib
import json
import os
import re
import shutil
import subprocess
import sys
import tempfile
import urllib2

import boto
import requests

import auth_headers

EVG_BASE = "https://evergreen.mongodb.com/rest/v1"
S3_BUCKET = "info-mongodb-com"
S3_PATH = "mongodb-bi/v2"
S3_DEV_RUN_BUCKET = "mciuploads"
S3_DEV_RUN_PATH = "sqlproxy/releases/mongodb-bi"
CURRENT_RELEASES_JSON = "current.json"
ARCHIVED_RELEASES_JSON = "full.json"
MAIN_DOWNLOADS_JSON = "mongodb-bi-downloads.json"
RELEASES_JSON = "mongodb-bi-releases.json"
UNITS = ['', 'Ki', 'Mi', 'Gi', 'Ti', 'Pi', 'Ei', 'Zi']
HASH_TYPES = ["md5", "sha1", "sha256"]
NUM_RELEASE_PLATFORMS = 29
ZIP_SUFFIX = " (zip)"
SIG_SUFFIX = " (sig)"
DEV_RUN = False
USAGE = """
BI Connector Release tool 0.1
Usage:
  release.py (-v VERSION | --version=VERSION)
  release.py (-h | --help)
  release.py (-d -v VERSION)
Options:
  -h --help     Show this screen.
  -v            Evergreen version identifier
  -d            Release files to development bucket
"""


def ensure_env():
    """Ensures the required environment variables are set.
    """
    if os.environ.get('EVG_USER', None) is None:
        print("Can't find 'EVG_USER' in environment variables")
        sys.exit(1)
    if os.environ.get('EVG_KEY', None) is None:
        print("Can't find 'EVG_KEY' in environment variables")
        sys.exit(1)
    if os.environ.get('AWS_ACCESS_KEY_ID', None) is None:
        print("Can't find 'AWS_ACCESS_KEY_ID' in environment variables")
        sys.exit(1)
    if os.environ.get('AWS_SECRET_ACCESS_KEY', None) is None:
        print("Can't find 'AWS_SECRET_ACCESS_KEY' in environment variables")
        sys.exit(1)

def run():
    """Runs the BI releaser.
    """
    global S3_BUCKET
    global S3_PATH
    global DEV_RUN

    version = ''
    try:
        opts, _ = getopt.getopt(sys.argv[1:], "hv:d", ["version="])
    except getopt.GetoptError:
        print(USAGE)
        sys.exit(2)
    for opt, arg in opts:
        if opt == '-h':
            print(USAGE)
            sys.exit(0)
        elif opt in ("-v", "--version"):
            version = arg
        if opt == '-d':
            print("Dev Run of release.py")
            S3_BUCKET = S3_DEV_RUN_BUCKET
            S3_PATH = S3_DEV_RUN_PATH
            DEV_RUN = True
    if version == '':
        print(USAGE)
        sys.exit(1)
    ensure_env()
    releaser = BIReleaser(version)
    releaser.run()

class BIReleaser(object):
    """Represents a release instance.
    """
    def __init__(self, version):
        """Initialize a new BIReleaser object.
        """
        self.__urls = {}
        self.__version = version
        self.__dev_release = False
        self.__release_version = ""
        self.__temp_dir = tempfile.mkdtemp()
        self.__bucket = boto.connect_s3().get_bucket(S3_BUCKET, validate=False)
        self.__headers = {
            'Api-User': os.environ.get('EVG_USER', None),
            'Api-Key': os.environ.get('EVG_KEY', None),
        }

    def download_binaries(self):
        """Downloads the binaries from Evergreen.
        """
        def sizeof_fmt(num, suffix='B'):
            """Formats bytes in human-readable fashion.
            """
            for unit in UNITS:
                if abs(num) < 1024.0:
                    return "%3.1f%s%s" % (num, unit, suffix)
                num /= 1024.0
            return "%.1f%s%s" % (num, 'Yi', suffix)

        for entry, url in self.__urls.items():
            file_name = url.split('/')[-1]
            data = urllib2.urlopen(url)
            file_path = open(os.path.join(self.__temp_dir, file_name), 'wb')
            meta = data.info()
            file_size = int(meta.getheaders("Content-Length")[0])
            print("Downloading '%s' binary: %s Bytes: %s" % \
             (entry, file_name, sizeof_fmt(file_size)))
            block_sz = 8192
            while True:
                buf = data.read(block_sz)
                if not buf:
                    break
                file_path.write(buf)
            file_path.close()

    def fetch_urls(self):
        """Fetches the release's binaries URL from the Evergreen API.
        """
        url = "%s/versions/%s" % (EVG_BASE, self.__version)

        print("Contacting Evergreen at %s" % (url))

        rpc = requests.get(url, headers=self.__headers)
        if rpc.status_code != 200:
            print("Can't contact Evergreen at %s: %s" % (url, rpc.text))
            rpc.raise_for_status()
        versions = json.loads(rpc.text)
        self.validate_release_commit(versions["message"])

        if self.__release_version == "":
            print("Will upload latest build for commit: %s" % (self.__version))

        builds = versions["builds"]

        if not builds:
            print("No builds found for version '%s'" % (self.__version))
            sys.exit(0)

        for build_name in builds:
            # don't upload binaries for race buildvariant
            if "race" in build_name:
                print("Skipping build %s" % (build_name))
                continue
            url = "%s/builds/%s" % (EVG_BASE, build_name)
            print("Fetching build %s" % (url))
            rpc = requests.get(url, headers=self.__headers)
            if rpc.status_code != 200:
                print("Can't contact Evergreen")
                rpc.raise_for_status()
            build = json.loads(rpc.text)
            sign = build["tasks"].get("sign", None)

            if sign is None:
                print("No sign task found for '%s'" % (build["name"]))
                continue

            status = sign.get("status", None)

            if status is None:
                print("No status found for '%s'" % (build["name"]))
                sys.exit(1)

            if status != "success":
                print("WARNING: %s sign task has status '%s': skipping..." % (build["name"], \
                    status))
                continue

            url = "%s/tasks/%s" % (EVG_BASE, sign["task_id"])
            print("Fetching task %s..." % (url))
            rpc = requests.get(url, headers=self.__headers)

            if rpc.status_code != 200:
                print("Can't contact Evergreen")
                rpc.raise_for_status()

            entry = json.loads(rpc.text)
            variant = entry["build_variant"]
            if "osx" in variant:
                extension = {".zip"}
            elif "windows" in variant:
                extension = {".msi", ".zip"}
            else:
                extension = {".tgz"}
            for entry_file in entry["files"]:
                url = entry_file["url"]
                _, ext = os.path.splitext(url)
                if ext in extension and self.__release_version in url:
                    variant_with_suffix = ''
                    if ext == ".zip" and "windows" in variant:
                        variant_with_suffix = variant + ZIP_SUFFIX
                    else:
                        variant_with_suffix = variant
                    self.__urls[variant_with_suffix] = url
                if ".sig" == ext and self.__release_version in url:
                    self.__urls[variant + SIG_SUFFIX] = url

        # *2 because we are uploading the sig file and an archive.
        # adding 1 since we upload the .zip binary for Windows.
        expected_artifact_count = NUM_RELEASE_PLATFORMS * 2 + 1
        if len(self.__urls) != expected_artifact_count:
            print("Expected %s URLs, got %s" % (expected_artifact_count, len(self.__urls)))
            sys.exit(1)

    def modify_version_string(self, file_name):
        """Modifies version string in file_name to be "latest" if
        it is present
        """
        # If the file ends with an extension followed by a checksum
        # extension (e.g. ".tgz.md5"), we must use a different regex
        # to match on the full extension.
        if any(file_name.endswith(hash_type) for hash_type in HASH_TYPES + ['sig']):
            match = re.match(r'(.*)-v.*\.(.*\..*)$', file_name)
        else:
            match = re.match(r'(.*)-v.*\.(.*)$', file_name)

        # match.group(1) is the part of the filename before -v
        # match.group(2) is the full file extension
        if match:
            file_name = "%s-latest.%s" % \
                (match.group(1), match.group(2))
        return file_name

    def run(self):
        """Runs an instance of the BI Releaser.
        """
        self.fetch_urls()
        self.download_binaries()
        self.upload_binaries()
        self.verify_website_links()
        self.write_releases_entry()
        self.upload_website_artifacts()

    def upload_binaries(self):
        """Uploads the release binaries to S3.
        """

        def create_and_upload_hash(key_name, file_name, file_path, hash_type):
            """Creates and uploads a checksum file for the given file and hash function.
            """
            # don't hash .sig files
            if file_path.endswith(".sig"):
                return

            hash_fxn = hashlib.new(hash_type)

            file_handle = open(file_path)
            read_file = file_handle.read()
            hash_fxn.update(read_file)
            hashed = hash_fxn.hexdigest()

            hash_file_name = "%s.%s" % (file_path, hash_type)
            hash_file = open(hash_file_name, "w+")
            # We use the same format as server.
            hash_file.write("%s  %s" % (hashed, file_name))
            hash_file.close()

            print("\t Uploading %s hash: %s" % (hash_type, hashed))

            key = self.__bucket.new_key("%s.%s" % (key_name, hash_type))
            key.set_contents_from_filename(hash_file_name)

        upload_path = S3_PATH
        if self.__release_version == "":
            upload_path = os.path.join(S3_PATH, "latest")

        for file_name in os.listdir(self.__temp_dir):
            print("Uploading binary %s..." % (file_name))
            key_name = os.path.join(upload_path, file_name)
            key = self.__bucket.new_key(key_name)
            file_location = os.path.join(self.__temp_dir, file_name)

            if self.__release_version == "":
                new_file_name = self.modify_version_string(file_name)
                if new_file_name is not file_name:
                    key_name = os.path.join(upload_path, new_file_name)
                    key = self.__bucket.new_key(key_name)

            key.set_contents_from_filename(file_location)

            for hash_type in HASH_TYPES:
                create_and_upload_hash(key_name, file_name, file_location, hash_type)

            if file_location.endswith(".zip") or file_location.endswith(".sig"):
                os.remove(file_location)

        # delete the zip archive from the URL map
        for key, _ in self.__urls.items():
            if key.endswith(ZIP_SUFFIX) or key.endswith(SIG_SUFFIX):
                del self.__urls[key]
                break

    def upload_current_artifacts(self):
        """Updates the main downloads page at
        https://www.mongodb.com/download-center#bi-connector
        and the current.json file at at:
        https://info-mongodb-com.s3.amazonaws.com/mongodb-bi/current.json
        """

        # read in main downloads template file
        with open(MAIN_DOWNLOADS_JSON, "r") as file_handle:
            new_release_version = file_handle.read()

        new_release_version = re.sub("S3_PATH", S3_PATH, new_release_version)
        new_release_version = re.sub("S3_BUCKET", S3_BUCKET, new_release_version)
        new_release_version = re.sub("RELEASE_VERSION", self.__release_version, new_release_version)

        for url, entry in self.__urls.items():
            new_release_version = re.sub(url.upper(), os.path.basename(entry), new_release_version)

        # download MAIN_DOWNLOADS_JSON file
        main_downloads_page = json.loads(
            self.__bucket.get_key(
                os.path.join(os.path.dirname(S3_PATH), MAIN_DOWNLOADS_JSON)).
            get_contents_as_string())

        release_version_components = self.__release_version.split(".")
        is_ga = len(release_version_components) == 3 and release_version_components[2] == "0"

        # the main downloads page should have, at most, 2 versions
        # available for download - 1 only, if it's GA
        if is_ga:
            # for GA, only the stable version should be available for
            # download on the main page
            main_downloads_page["versions"] = []
        else:
            delete_index = -1
            # find and remove older version from main release page
            for i, main_version in enumerate(main_downloads_page["versions"]):
                current_version_components = main_version["version"].split(".")
                if int(current_version_components[1]) == int(release_version_components[1]):
                    delete_index = i
                    break

            # delete any older version found
            if delete_index != -1:
                del main_downloads_page["versions"][delete_index]

        # add new version to MAIN_DOWNLOADS_JSON
        main_downloads_page["versions"].append(json.loads(new_release_version))
        self.update_json_file(MAIN_DOWNLOADS_JSON, main_downloads_page)

        # download CURRENT_RELEASES_JSON file
        current_releases_page = json.loads(
            self.__bucket.get_key(
                os.path.join(os.path.dirname(S3_PATH), CURRENT_RELEASES_JSON)).
            get_contents_as_string())

        delete_index = -1

        # find and remove older version from current release page
        for i, version in enumerate(current_releases_page["versions"]):
            current_version_components = version["version"].split(".")
            if int(current_version_components[0]) == int(release_version_components[0]) and \
                int(current_version_components[1]) == int(release_version_components[1]):
                delete_index = i
                break

        # delete any older version found
        if delete_index != -1:
            del current_releases_page["versions"][delete_index]

        # read in templated release information
        with open(os.path.join(self.__temp_dir, RELEASES_JSON), "r") as handle:
            new_release_version = json.loads("".join(handle.readlines()))

        # add new current release information
        current_releases_page["versions"].append(new_release_version)
        self.update_json_file(CURRENT_RELEASES_JSON, current_releases_page)

    def upload_full_artifacts(self):
        """Updates the ARCHIVED_RELEASES_JSON file at:
        https://info-mongodb-com.s3.amazonaws.com/mongodb-bi/full.json
        """

        # download full release page JSON file
        key_name = os.path.join(os.path.dirname(S3_PATH), ARCHIVED_RELEASES_JSON)
        key = self.__bucket.get_key(key_name)
        archived_downloads_page = json.loads(key.get_contents_as_string())

        release_version_components = []
        for i in self.__release_version.split("."):
            if i.isdigit():
                release_version_components.append(ast.literal_eval(i))
            else:
                release_version_components.append(i)

        # find and remove older version from full release page
        duplicates = []
        for version in archived_downloads_page["versions"]:
            components = []
            for i in version["version"].split("."):
                if i.isdigit():
                    components.append(ast.literal_eval(i))
                else:
                    components.append(i)
            if int(components[0]) == int(release_version_components[0]) and \
                int(components[1]) == int(release_version_components[1]):
                version["current"] = False
                if components[2] == release_version_components[2]:
                    duplicates.append(version)

        for entry in duplicates:
            archived_downloads_page["versions"].remove(entry)

        # read in templated release information
        with open(os.path.join(self.__temp_dir, RELEASES_JSON), "r") as handle:
            new_release = json.loads("".join(handle.readlines()))

        # add new archive release information
        archived_downloads_page["versions"].append(new_release)
        self.update_json_file(ARCHIVED_RELEASES_JSON, archived_downloads_page)

    def update_json_file(self, name, data):
        """Updates the named S3 JSON file with data.
        """
        def extract_version(obj):
            """Extracs the version key from the json document or returns 0 if the key doesn't exist.
            """
            try:
                return obj['version']
            except KeyError:
                return 0
        data["versions"].sort(key=extract_version, reverse=True)

        tmp = os.path.join(self.__temp_dir, "tmp")
        with open(tmp, "w") as file_handle:
            json.dump(data, file_handle, sort_keys=True, indent=4)

        print("Updating %s page for BI Connector" % (name))

        key_name = os.path.join(os.path.dirname(S3_PATH), name)
        self.__bucket.copy_key(key_name+".bk", S3_BUCKET, key_name)
        self.__bucket.new_key(key_name).set_contents_from_filename(
            os.path.join(self.__temp_dir, tmp))

    def upload_website_artifacts(self):
        """Determines what release artifacts to update and upload.
        """
        self.upload_full_artifacts()
        if not self.__dev_release:
            self.upload_current_artifacts()
        shutil.rmtree(self.__temp_dir)

    def validate_release_commit(self, message):
        """Ensures the commit message meets the release format.
        """
        if not message.startswith("BUMP"):
            print("Commit does not start with 'BUMP': releasing nightly")
            return

        parts = message.split(" ")
        if len(parts) != 2:
            print("Commit does not contain release version: releasing nightly")
            return

        version = parts[1].strip()

        # strip v in commit message
        if version.startswith("v"):
            version = version[1:]

        if version.count(".") != 2:
            print("Commit contains invalid version '%s': releasing nightly" % (version))
            return

        tmp = version

        if "-" in version:
            tmp = version[0:version.index("-")]

        if [y for y in tmp.split(".") if not y.isdigit()]:
            print("Commit contains non-numeric version '%s': releasing nightly" % (version))
            return

        self.__release_version = version

    def verify_website_links(self):
        """Verifies that the download URLs exist and are valid.
        """
        for file_name in os.listdir(self.__temp_dir):
            # If this version is not an official release, files are placed in a
            # "latest" subdirectory on s3.
            latest = ""
            if self.__release_version == "":
                latest = "latest/"
                file_name = self.modify_version_string(file_name)

            host = "%s.s3.amazonaws.com" % S3_BUCKET
            canonical_uri = "/%s/%s%s" % (S3_PATH, latest, file_name)

            url = "https://%s%s"  % (host, canonical_uri)

            # Because the development S3 bucket is private,
            # we must make a request with authorization headers to visit the links.
            rpc = None
            if DEV_RUN:
                headers = auth_headers.construct_headers_for_head(host, canonical_uri)
                rpc = requests.head(url, headers=headers)
            else:
                rpc = requests.head(url)

            if rpc.status_code == 200:
                print('%s URL is fine...' % (url))
            else:
                print('%s URL returned %s' % (url, rpc.status_code))
                rpc.raise_for_status()

        # for nightly uploads, don't update any links
        if self.__release_version == "":
            print("Finished uploading nightly build")
            sys.exit(0)

    def write_releases_entry(self):
        """Writes new release information to file.
        """
        try:
            revision = subprocess.Popen(
                "git rev-list -n 1 v%s" % (self.__release_version),
                shell=True, stdout=subprocess.PIPE).stdout.read().strip("\n")
        except subprocess.CalledProcessError as err:
            print(err)
            sys.exit(1)

        is_dev_release = "false"
        is_prod_release = "true"
        is_rc_candidate = "false"

        if self.__dev_release:
            is_dev_release = "true"
            is_prod_release = "false"

        if "-" in self.__release_version:
            is_rc_candidate = "true"

        notes_anchor = self.__release_version.replace(".", "-", -1)

        # read in current release template
        with open(RELEASES_JSON, "r") as file_handle:
            contents = file_handle.read()

        # update new release information in template file
        contents = re.sub("S3_PATH", S3_PATH, contents)
        contents = re.sub("S3_BUCKET", S3_BUCKET, contents)
        contents = re.sub("GITHASH", revision, contents)
        contents = re.sub("IS_DEV_RELEASE", is_dev_release, contents)
        contents = re.sub("IS_RC_CANDIDATE", is_rc_candidate, contents)
        contents = re.sub("IS_PROD_RELEASE", is_prod_release, contents)
        contents = re.sub("NOTES_ANCHOR", notes_anchor, contents)
        contents = re.sub("RELEASE_VERSION", self.__release_version, contents)
        contents = re.sub("DATE", str(datetime.date.today()), contents)

        for url, entry in self.__urls.items():
            contents = re.sub(url.upper(), os.path.basename(entry), contents)

        new_releases_json = os.path.join(self.__temp_dir, RELEASES_JSON)

        with open(new_releases_json, "w") as file_handle:
            file_handle.write(contents)

        data = json.loads(contents)

        # final validity check
        valid = True

        for url in self.__urls:
            if url.upper() in data:
                valid = False
                print("Could not find binary for %s release" % (url))
        if not valid:
            sys.exit(1)

        print("Release file passed validation check")


if __name__ == '__main__':
    run()
