[
	{
		"$match" : {
			"query.plan.fully_pushed_down" : false
		}
	},
	{
		"$unwind" : {
			"path" : "$query.plan.stages"
		}
	},
	{
		"$match" : {
			"query.plan.stages.pushdown_errors" : {
				"$exists" : true
			}
		}
	},
	{
		"$unwind" : {
			"path" : "$query.plan.stages.pushdown_errors"
		}
	},
	{
		"$match" : {
			"query.plan.stages.pushdown_errors.reason" : {
				"$ne" : "unable to push down source stage"
			}
		}
	},
	{
		"$project" : {
			"_id" : 0,
			"sql" : "$query.sql",
			"stage" : "$query.plan.stages.stage_type",
			"pushdown_blocker" : "$query.plan.stages.pushdown_errors.name",
			"pushdown_error" : "$query.plan.stages.pushdown_errors.reason"
		}
	},
	{
		"$group" : {
			"_id" : "$pushdown_error",
			"count" : {
				"$sum" : 1
			},
			"queries_affected" : {
				"$addToSet" : "$sql"
			},
			"pushdown_blockers" : {
				"$addToSet" : "$pushdown_blocker"
			}
		}
	},
	{
		"$sort" : {
			"count" : -1
		}
	}
]

