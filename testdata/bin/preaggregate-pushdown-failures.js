use metrics;

const current_date = new Date();
const current_year = current_date.getUTCFullYear();
const current_month = current_date.getUTCMonth()+1;
const current_month_expiration_date_str = `${current_year+2}-${current_month<=9?'0':''}${current_month}-01T12:00:00Z`;
const current_month_expiration_date = new Date(current_month_expiration_date_str);

const pipeline = [
	{'$match': {
		'query.plan.fully_pushed_down': false,
		'expire_at': current_month_expiration_date,
	}},
	{'$unwind': '$query.plan.stages'},
	{'$unwind': {
		'path': '$query.plan.stages.pushdown_errors',
		'includeArrayIndex': 'failure_reason_idx',
	}},
	{'$project': {
		'_id': {
			'query_id': '$_id',
			'stage_id': '$query.plan.stages.id',
			'reason_idx': '$failure_reason_idx',
		},
		'sql': '$query.sql',
		'mongodb_version': '$variables.mongodb_version',
		'mongosqld_version': '$variables.mongosqld_version',
		'pushdown_failure': {
			'stage_type': '$query.plan.stages.stage_type',
			'blocker': '$query.plan.stages.pushdown_errors.name',
			'reason': '$query.plan.stages.pushdown_errors.reason',
		},
	}},
	{'$merge': {
		'into': 'atlas_bic_pushdown_failure_preagg',
		'whenMatched': 'keepExisting',
		'whenNotMatched': 'insert',
	}},
];

print(`starting pre-aggregation at ${new Date()}`);
db.atlas_bic.aggregate(pipeline);
print(`pre-aggregation completed successfully at ${new Date()}`);
