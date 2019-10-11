use metrics;

const current_date = new Date();
const current_year = current_date.getUTCFullYear();
const current_month = current_date.getUTCMonth()+1;
const last_month = parseInt(current_month === 1 ? 12 : current_month-1, 10);
const last_month_year = parseInt(last_month === 12 ? current_year-1 : current_year, 10);

const current_month_expiration_date_str = `${current_year+2}-${current_month<=9?'0':''}${current_month}-01T12:00:00Z`;
const current_month_expiration_date = new Date(current_month_expiration_date_str);

const last_month_expiration_date_str = `${last_month_year+2}-${last_month<=9?'0':''}${last_month}-01T12:00:00Z`;
const last_month_expiration_date = new Date(last_month_expiration_date_str);

const rollupsThisMonthPipeline = [
	{'$project': {
		'year': {'$year': '$started'},
		'month': {'$month': '$started'},
	}},
	{'$match': {
		'year': current_year,
		'month': current_month,
	}},
];

const createRollupPipeline = (rollup_type) => {
	let expires;
	let date;

	if (rollup_type === 'full_month') {
		expires = last_month_expiration_date;
		date = {
			'month': { '$literal': last_month },
			'year': { '$literal': last_month_year },
		};
	} else if (rollup_type === 'month_to_date') {
		expires = current_month_expiration_date;
		date = {
			'day': {'$dayOfMonth': current_date},
			'month': {'$month': current_date},
			'year': {'$year': current_date},
		};
	} else {
		throw `unexpected rollup type "${rollup_type}"`;
	};

	// The pipeline below uses $facet instead of $group because we
	// want to generate a single document to be inserted into the
	// atlas_bic_rollups collection. While the current aggregation
	// could be written as a $group and then assembled into a
	// document after it returns, we may want to add other
	// aggregates to these rollups in the future that can't be
	// expressed with a single group.
	return [
		{'$match': {'expire_at': expires}},
		{'$facet' : {
			'stats_month': [
				{'$match': {'expire_at': expires}},
				{'$addFields': {'stage_count': {'$size': '$query.plan.stages'}}},
				{'$group': {
					'_id': {
						'mongodb_version': '$variables.mongodb_version',
						'mongosqld_version': '$variables.mongosqld_version',
					},
					'query_count': {'$sum': 1},
					'pushdown_failure_count': {'$sum': {'$cond': { 'if': '$query.plan.fully_pushed_down', 'then': 0, 'else': 1 }}},
					'latency_avg': {'$avg': '$query.execution.latency_ms'},
					'latency_max': {'$max': '$query.execution.latency_ms'},
					'latency_stddev': {'$stdDevPop': '$query.execution.latency_ms'},
					'stage_count_avg': {'$avg': '$stage_count'},
					'stage_count_max': {'$max': '$stage_count'},
					'stage_count_stddev': {'$stdDevPop': '$stage_count'},
				}},
			],
		}},
		{'$project' : {
			'rollup_type': rollup_type,
			'date': date,
			'started': current_date,
			[rollup_type] : {
				'queries' : {
					'total': {'$sum': '$stats_month.query_count'},
					'not_fully_pushed_down': {'$sum': '$stats_month.pushdown_failure_count'},
				},
				'stats_by_version': "$stats_month",
			},
		}},
	];
};

const numRollupsThisMonth = db.atlas_bic_rollups.aggregate(rollupsThisMonthPipeline).toArray().length;
const firstOfMonth = numRollupsThisMonth == 0;

const rollupPipeline = function() {
	if (firstOfMonth) {
		print('first rollup of the month');
		print('performing monthly rollup');
		return createRollupPipeline('full_month');
	} else {
		print('not the first rollup of the month');
		print('performing incremental rollup');
		return createRollupPipeline('month_to_date');
	}
}();

print('running aggregation');

let rollup = db.atlas_bic.aggregate(rollupPipeline).toArray();
if (rollup.length !== 1) {
	throw `expected on document, but got ${rollup.length}`;
}
rollup = rollup[0];

print('aggregation completed successfully');
print('inserting rollup doc into atlas_bic_rollups');

rollup.finished = new Date();
db.atlas_bic_rollups.insertOne(rollup);
