"""
This script supports the release process for the BI Connector.
"""

import ast
import datetime
import json
import os
import tempfile
import shutil
import subprocess
import sys
import urllib2
import requests
import docopt
import boto
import httplib2

EVG_BASE = "https://evergreen.mongodb.com/rest/v1"
S3_BUCKET = "info-mongodb-com"
S3_PATH = "mongodb-bi/v2"
CURRENT_JSON = "current.json"
FULL_JSON = "full.json"
MAIN_FILE = "mongodb-bi-downloads.json"
RELEASES_FILE = "mongodb-bi-releases.json"
UNITS = ['', 'Ki', 'Mi', 'Gi', 'Ti', 'Pi', 'Ei', 'Zi']
USAGE = """
BI Connector Release tool 0.1
Usage:
  release.py (-v VERSION | --version=VERSION)
  release.py (-h | --help)
Options:
  -h --help     Show this screen.
  -v            Evergreen version identifier
"""

def ensure_env():
    """Ensures the required environment variables are set.
    """
    if os.environ.get('EVG_USER', None) is None:
        print("Can't find Evergreen credentials in 'EVG_USER' environment \
        variable. Fetch from https://evergreen.mongodb.com/settings")
        sys.exit(1)
    if os.environ.get('EVG_KEY', None) is None:
        print("Can't find Evergreen credentials in 'EVG_KEY' environment \
        variable. Fetch from https://evergreen.mongodb.com/settings")
        sys.exit(1)
    if os.environ.get('AWS_ACCESS_KEY_ID', None) is None:
        print("Can't find Evergreen credentials in 'AWS_ACCESS_KEY_ID' \
        environment variable")
        sys.exit(1)
    if os.environ.get('AWS_SECRET_ACCESS_KEY', None) is None:
        print("Can't find Evergreen credentials in 'AWS_SECRET_ACCESS_KEY' \
            environment variable")
        sys.exit(1)

class BIReleaser(object):
    """Represents a release instance.
    """
    def __init__(self, opts):
        """Initialize a new BIReleaser object.
        """
        self.__urls = {}
        self.__version = opts["--version"]
        self.__dev_release = False
        self.__release_candidate = False
        self.__release_version = ""
        self.__temp_dir = tempfile.mkdtemp()
        self.__bucket = boto.connect_s3().get_bucket(S3_BUCKET)

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
            file_size_dl = 0
            block_sz = 8192
            while True:
                buf = data.read(block_sz)
                if not buf:
                    break
                file_size_dl += len(buf)
                file_path.write(buf)
                status = r"%10d  [%3.2f%%]" % \
                (file_size_dl, file_size_dl * 100. / file_size)
                status = status + chr(8)*(len(status)+1)
                print status,
            file_path.close()

    def fetch_urls(self):
        """Fetches the release's binaries URL from the Evergreen API.
        """
        url = "%s/versions/%s" % (EVG_BASE, self.__version)
        user_name = os.environ.get('EVG_USER', None)
        evg_key = os.environ.get('EVG_KEY', None)
        headers = {'Auth-Username': user_name, 'Api-Key': evg_key}
        print("Contacting Evergreen at %s" % (url))
        response = requests.get(url, headers=headers)
        if response.status_code != 200:
            print("Can't contact Evergreen at %s: %s" % (url, response.text))
            sys.exit(1)
        versions = json.loads(response.text)
        builds = versions["builds"]

        if len(builds) == 0:
            print "No builds found for version '%s'" % (self.__version)
            sys.exit(0)

        for build_name in builds:
            # don't upload binaries for race buildvariant
            if "race" in build_name:
                print("Skipping build %s..." % (build_name))
                continue
            url = "%s/builds/%s" % (EVG_BASE, build_name)
            print("Fetching build %s..." % (url))
            response = requests.get(url, headers=headers)
            if response.status_code != 200:
                print("Can't contact Evergreen")
                response.raise_for_status()
            build = json.loads(response.text)
            task = build["tasks"]
            dist = task.get("dist", None)

            if dist is None:
                print("No dist task found for '%s'" % (build["name"]))
                continue

            status = dist.get("status", None)

            if status is None:
                print("No status found for '%s'" % (build["name"]))
                sys.exit(1)

            if status != "success":
                print("%s dist has status '%s', exiting..." % (build["name"], \
                    status))
                sys.exit(1)

            url = "%s/tasks/%s" % (EVG_BASE, dist["task_id"])
            print("Fetching task %s..." % (url))
            response = requests.get(url, headers=headers)

            if response.status_code != 200:
                print("Can't contact Evergreen")
                sys.exit(1)

            entry = json.loads(response.text)
            url = entry["files"][0]["url"]
            variant = entry["build_variant"]
            self.__urls[variant] = url

    def prompt(self):
        """Prompts the user for the release version.
        """

        try:
            tag = subprocess.Popen("git describe",
                shell=True,stdout=subprocess.PIPE).stdout.read().strip("\n")
            if tag.startswith("v"):
                tag = tag[1:]
        except subprocess.CalledProcessError as err:
            print err
            sys.exit(1)

        while True:
            resp = raw_input('Release Version [%s]): ' % (tag)) or tag

            if resp.count(".") != 2:
                print("Invalid Release Version")
            else:
                tmp = resp

                if "-" in resp:
                    tmp = resp[0:resp.index("-")]

                if len([y for y in tmp.split(".") if not y.isdigit()]) != 0:
                    print("Invalid Release Version")
                    continue

                break

        self.__release_version = resp

        if "-beta" in tmp:
            self.__dev_release = True
        if "-rc" in tmp:
            self.__release_candidate = True

    def run(self):
        """Runs an instance of the BI Releaser.
        """
        self.prompt()
        self.fetch_urls()
        self.download_binaries()
        self.upload_binaries()
        self.verify_website_links()
        self.write_releases_entry()
        self.upload_website_artifacts()

    def upload_binaries(self):
        """Uploads the release binaries to S3.
        """
        for file_name in os.listdir(self.__temp_dir):
            print "Uploading binary %s..." % (file_name)
            key_name = os.path.join(S3_PATH, file_name)
            key = self.__bucket.new_key(key_name)
            key.set_contents_from_filename(
                os.path.join(self.__temp_dir, file_name))

    def upload_current_artifacts(self):
        """Updates the main downloads page at
        https://www.mongodb.com/download-center#bi-connector
        and the current.json file at at:
        https://info-mongodb-com.s3.amazonaws.com/mongodb-bi/current.json
        """
        # read in main downloads template file
        with open(MAIN_FILE, "r") as file_handle:
            template = file_handle.readlines()

        new_file = os.path.join(self.__temp_dir, MAIN_FILE)

        with open(new_file, "w") as file_handle:
            for line in template:
                line = line.replace(
                    "S3_PATH", S3_PATH, -1).replace(
                        "S3_BUCKET", S3_BUCKET, -1).replace(
                            "VERSION", self.__release_version, -1)
                for url, entry in self.__urls.items():
                    basename = os.path.basename(entry)
                    line = line.replace(url.upper(), basename, -1)
                file_handle.write(line)

        print("Updating main downloads page for BI Connector")

        # make a backup first
        key_name = os.path.join(os.path.dirname(S3_PATH), MAIN_FILE)
        self.__bucket.copy_key(key_name+".bk", S3_BUCKET, key_name)
        self.__bucket.new_key(key_name).set_contents_from_filename(new_file)

        # download current release page JSON file
        key = self.__bucket.get_key(
            os.path.join(os.path.dirname(S3_PATH), CURRENT_JSON))
        current = json.loads(key.get_contents_as_string())
        new_component = self.__release_version.split(".")
        delete_index = -1

        # find and remove older version from current release page
        for i, version in enumerate(current["versions"]):
            components = version["version"].split(".")
            if int(components[0]) == int(new_component[0]) and \
                int(components[1]) == int(new_component[1]):
                delete_index = i
                break

        # delete any older version found
        if delete_index != -1:
            del current["versions"][delete_index]

        # read in templated release information
        with open(os.path.join(self.__temp_dir, RELEASES_FILE), "r") as handle:
            new_release = json.loads("".join(handle.readlines()))

        # add new current release information
        current["versions"].append(new_release)
        self.update_json_file(CURRENT_JSON, current)

    def upload_full_artifacts(self):
        """Updates the full.json file at at:
        https://info-mongodb-com.s3.amazonaws.com/mongodb-bi/full.json
        """

        # download full release page JSON file
        key = self.__bucket.get_key(os.path.join(
            os.path.dirname(S3_PATH), FULL_JSON))
        full = json.loads(key.get_contents_as_string())
        new_component = []
        for i in self.__release_version.split("."):
            if i.isdigit():
                new_component.append(ast.literal_eval(i))
            else:
                new_component.append(i)

        # find and remove older version from full release page
        duplicates = []
        for version in full["versions"]:
            components = []
            for i in version["version"].split("."):
                if i.isdigit():
                    components.append(ast.literal_eval(i))
                else:
                    components.append(i)
            if int(components[0]) == int(new_component[0]) and \
                int(components[1]) == int(new_component[1]):
                version["current"] = False
                if components[2] == new_component[2]:
                    duplicates.append(version)

        for entry in duplicates:
            full["versions"].remove(entry)

        # read in templated release information
        with open(os.path.join(self.__temp_dir, RELEASES_FILE), "r") as handle:
            new_release = json.loads("".join(handle.readlines()))

        # add new full release information
        full["versions"].append(new_release)
        self.update_json_file(FULL_JSON, full)

    def update_json_file(self, name, data):
        """Updates the named S3 JSON file with data.
        """
        def extract_version(json):
            try:
                return json['version']
            except KeyError:
                return 0
        data["versions"].sort(key=extract_version, reverse=True)

        tmp = os.path.join(self.__temp_dir, "tmp")
        with open(tmp, "w") as file_handle:
            json.dump(data, file_handle, sort_keys=True, indent=4)

        print "Updating %s page for BI Connector" % (name)

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

    def verify_website_links(self):
        """Verifies that the download URLs exist and are valid.
        """
        http_handle = httplib2.Http()
        for url, entry in self.__urls.items():
            request = http_handle.request(entry, 'HEAD')
            if request[0]['status'] == "200":
                print('%s URL is fine...' % (url))
            else:
                print('%s URL returned %s' % (url, request[0]['status']))
                sys.exit(1)

    def write_releases_entry(self):
        """Writes new release information to file.
        """
        # read in current release template
        with open(RELEASES_FILE, "r") as file_handle:
            template = file_handle.readlines()

        new_releases_file = os.path.join(self.__temp_dir, RELEASES_FILE)

        try:
            revision = subprocess.Popen(
                "git rev-list -n 1 v%s" % (self.__release_version),
                shell=True,stdout=subprocess.PIPE).stdout.read().strip("\n")
        except subprocess.CalledProcessError as err:
            print err
            sys.exit(1)

        dev_release = "false"
        prod_release = "true"
        rc_candidate = "false"

        if self.__dev_release:
            dev_release = "true"
            prod_release = "false"

        if "-" in self.__release_version:
            rc_candidate = "true"

        notes_anchor = self.__release_version.replace(".", "-", -1)

        # update new release information in template file
        with open(new_releases_file, "w") as file_handle:
            for line in template:
                line = line.replace(
                    "S3_PATH", S3_PATH, -1).replace(
                        "GITHASH", revision, -1).replace(
                            "S3_BUCKET", S3_BUCKET, -1).replace(
                                "DEV-RELEASE", dev_release, -1)
                line = line.replace(
                    "RC-CANDIDATE", rc_candidate, -1).replace(
                        "PROD-RELEASE", prod_release, -1).replace(
                            "NOTES_ANCHOR", notes_anchor, -1).replace(
                                "VERSION", self.__release_version, -1).replace(
                                    "DATE", str(datetime.date.today()), -1)

                for url, entry in self.__urls.items():
                    basename = os.path.basename(entry)
                    line = line.replace(url.upper(), basename, -1)

                file_handle.write(line)

        with open(new_releases_file, "r") as file_handle:
            data = json.loads("".join(file_handle.readlines()))

        # final validity check
        valid = True
        for url in self.__urls:
            if url.upper() in data:
                valid = False
                print("Could not find binary for %v release" % (url))
        if not valid:
            sys.exit(1)
        print("Release file passed validation check")

if __name__ == '__main__':
    options = docopt.docopt(USAGE)
    ensure_env()
    releaser = BIReleaser(options)
    releaser.run()

