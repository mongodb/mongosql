# Examples

The easiest mode in which to use the mongodb-schema-manager is connecting to a local cluster with
username and password:

```
mongodb-schema-manager --uri mongodb://username:password@localhost
```

The mongodb-schema-manager will output the schemata namespaces generated:

```
Schema creation has completed successfully.
Now printing all namespaces with created or modified schemas:
Database: foo2. Namespaces: ["coll"].
Database: oncall. Namespaces: ["personneljob"].
Database: zugzugs. Namespaces: ["zugzug_result_13552373-6308-4972-a277-d6fa0ed8f81e", "zugzug_result_4bee6bff-5ccb-41ad-96d4-c96ea43205de", "zugzug_result_92d8e06a-c538-4b98-b83b-a25c24bce201", "zugzug_result_cc8ffd89-afda-49a9-9fd1-705e6574f367", "zugzugs"].
```

Alternatively, one can specify a config file:

```
mongodb-schema-manager -f ./config.yml
```

an example of a valid config file can be seen [here](./valid_config.yml)

Using Azure IMDS oidc for workload can be achieved as follows:

```
mongodb-schema-manager --uri mongodb://somecluster.a.query.mongodb.net/?ssl=true&authMechanism=MONGODB-OIDC&authMechanismProperties=ENVIRONMENT:azure,TOKEN_RESOURCE:http://example.com
```

The ns-include and ns-exclude options can be used to include and exclude namespaces as desired:

```
mongodb-schema-manager --uri mongodb://localhost --ns-include zugzugs.* --ns-exclude *.zugzug_result_92d8e06a-c538-4b98-b83b-a25c24bce201
```

Results in:

```
Schema creation has completed successfully.
Now printing all namespaces with created or modified schemas:
Database: zugzugs. Namespaces: ["zugzug_result_13552373-6308-4972-a277-d6fa0ed8f81e", "zugzug_result_4bee6bff-5ccb-41ad-96d4-c96ea43205de", "zugzug_result_cc8ffd89-afda-49a9-9fd1-705e6574f367", "zugzugs"].
```
