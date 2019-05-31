exports = function(){
  /*
    Accessing application's values:
    var x = context.values.get("value_name");

    Accessing a mongodb service:


    To call other named functions:
    var result = context.functions.execute("function_name", arg1, arg2);

    Try running in the console below.
  */
  var metrics = context.services.get("mongodb-atlas").db("metrics");
  var atlas_bic = metrics.collection("atlas_bic");
  var testCol = metrics.collection("test_trigger_col");
  
  atlas_bic.findOne().then(function(doc) {
    var msg = "document created by stitch scheduled trigger";
  
    testCol.insertOne({
      created: new Date(),
      message: msg,
      foundQuery: doc.query.sql,
    });
  });

  return {success: true};
};
