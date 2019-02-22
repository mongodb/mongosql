[
	{
		"$unwind" : {
			"path" : "$query.meta.functions"
			}
	},
	{
		"$match" : {
			"query.meta.functions.name" : {
				"$exists" : true
			}
		}
	},
	{
		"$group" : {
			"_id" : "$query.meta.functions.name",
			"latency_ms" : {
				"$avg" : "$query.execution.latency_ms"
			},
			"queries_affected" : {
				"$addToSet" : "$query.sql"
			}
		}
	},
	{
		"$sort" : {
			"count" : -1
		}
	}
]
