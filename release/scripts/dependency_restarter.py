"""
This script restarts the latest versions of the projects dependent on the BI Connector.
"""

import os
import json
import sys

import requests

EVG_BASE_V1 = "https://evergreen.mongodb.com/rest/v1"
EVG_BASE_V2 = "https://evergreen.mongodb.com/rest/v2"
DEPENDENT_PROJECTS = ["mongo-odbc-driver", "mongosql-auth-c"]


def ensure_env():
    """Ensures the required environment variables are set.
    """
    if os.environ.get('EVG_USER', None) is None:
        print("Can't find 'EVG_USER' in environment variables")
        sys.exit(1)
    if os.environ.get('EVG_KEY', None) is None:
        print("Can't find 'EVG_KEY' in environment variables")
        sys.exit(1)

def run():
    """Runs the dependency restarter.
    """
    dependency_restarter = DependencyRestarter()
    dependency_restarter.get_latest_versions()
    dependency_restarter.restart_latest_versions()

class DependencyRestarter(object):
    """Represents a dependency restarter instance.
    """
    def __init__(self):
        """Initialize a new DependencyRestarter object.
        """
        self.__versions = []
        self.__headers = {
            'Api-User': os.environ.get('EVG_USER', None),
            'Api-Key': os.environ.get('EVG_KEY', None),
        }

    def get_latest_versions(self):
        """Fetches the latest versions fron the dependent repositories.
        """
        for project in DEPENDENT_PROJECTS:
            url = "%s/projects/%s/versions" % (EVG_BASE_V1, project)
            print("Contacting Evergreen at %s" % (url))

            rpc = requests.get(url, headers=self.__headers)

            if rpc.status_code != 200:
                print("Can't contact Evergreen at %s: %s" % (url, rpc.text))
                rpc.raise_for_status()
            response = json.loads(rpc.text)

            versions = response.get('versions', None)
            if not versions:
                print("'versions' key not found for '%s'" % (project))
                sys.exit(1)

            if not versions:
                print ("No versions found for '%s'" % (project))
                sys.exit(1)

            latest_version = versions[0].get('version_id')

            if not latest_version:
                print("'version_id' key not found for '%s'" % (project))
                sys.exit(1)

            print("Latest version for %s is %s" % (project, latest_version))

            self.__versions.append(latest_version)

    def restart_latest_versions(self):
        """Restarts the latest versions in the dependent repositories.
        """
        for version in self.__versions:
            url = "%s/versions/%s/restart" % (EVG_BASE_V2, version)
            print("Restarting version at %s" % (version))

            rpc = requests.post(url, headers=self.__headers, data={})

            if rpc.status_code != 200:
                print("Can't contact Evergreen at %s: %s" % (url, rpc.text))
                rpc.raise_for_status()
            response = json.loads(rpc.text)

            print("Version %s ('%s') is now restarted" % (version, response.get('message', "???")))


if __name__ == '__main__':
    ensure_env()
    run()
