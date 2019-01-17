[
	{
		"$unwind" : {
			"path" : "$query.meta.unions",
			"preserveNullAndEmptyArrays" : true
		}
	},
	{
		"$group" : {
			"_id" : "$query.meta.unions.kind",
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
					"no_unions"
				]
			},
			"count" : 1
		}
	}
]
