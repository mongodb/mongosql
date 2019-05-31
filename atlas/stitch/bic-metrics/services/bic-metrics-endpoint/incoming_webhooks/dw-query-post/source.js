exports = function(payload, response) {
  // Connect to MongoDB Atlas
  var atlas = context.services.get('mongodb-atlas');
  var queries = atlas.db('metrics').collection('dw_bic');
  
  // Parse the stringified JSON body into an EJSON object
  var queryRecordsString = payload.body.text();
  var queryRecords = EJSON.parse(queryRecordsString);
  
  queries.insertMany(queryRecords).then(a => {
    // If all went according to plan, return a response object
    var res = { message: "successfully inserted records", records: queryRecords };
    response.setStatusCode(201);   // 201 - Resource Created
    response.setBody(JSON.stringify(res)); // Response body is the order document that was just inserted
  });
};