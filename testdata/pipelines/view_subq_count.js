[
	{
		"$unwind" : {
			"path" : "$query.meta.subqueries",
			"preserveNullAndEmptyArrays" : true
		}
	},
	{
		"$group" : {
			"_id" : "$query.meta.subqueries.kind",
			"count" : {
				"$sum" : 1
			}
		}
	},
	{
		"$project" : {
			"_id" : {
				"$ifNull" : [
					"$_id",
					"no_subqueries"
				]
			},
			"count" : 1
		}
	}
]
