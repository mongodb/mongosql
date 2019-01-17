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
			"count" : {
				"$sum" : "$query.meta.functions.count"
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
