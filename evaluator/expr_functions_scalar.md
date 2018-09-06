# Scalar Functions FAQ
#### last updated by Kaitlin Mahar, November 2017


### What is a scalar function?
MySQL has a number of scalar functions: functions that take in 0+ values from a single row and return a single result for each row.  This is opposed to aggregate functions, which aggregate values from multiple rows, like `avg` or `sum`. Some examples:
* [`lower(str)`](https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_lower)
* [`log(x)`](https://dev.mysql.com/doc/refman/5.7/en/mathematical-functions.html#function_log)
* [`left(str, len)`](https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_left)

### How are they used?
You can use scalar functions with both _literal_ arguments and _expressions_. A literal is a value you simply type in. For example, 5 is a literal value here:  
```
mysql> select log(5);
+--------------------+
| log(5)             |
+--------------------+
| 1.6094379124341003 |
+--------------------+
```
On the other hand, an expression refers to a SQL expression. This could be a reference to a column, or even the result of other MySQL functions and operators. Some examples - these are nonsense calculations, but you get the point - 
```
mysql> select log(c) from bar;
+--------------------+
| log(c)             |
+--------------------+
| 0.3756929497744942 |
|               NULL |			c is a column in bar
|  4.038377203500196 |
|               NULL |
|               NULL |
+--------------------+

mysql> select log(ln(c)) from bar; 
+-------------------+
| log(ln(c))        |
+-------------------+
| -0.97898309215057 |
|              NULL |			ln is another scalar function
| 1.395842928975181 |
|              NULL |
|              NULL |
+-------------------+

mysql> select log(pow(c, id)) from bar;
+-------------------+
| log(pow(c, id))   |
+-------------------+
|                 0 |
|              NULL |		pow is another scalar function
| 8.076754407000392 |
|              NULL |
|              NULL |
+-------------------+

mysql> select log(c + 1) from bar; 
+-------------------+
| log(c+1)          |
+-------------------+
| 0.898534010284896 |
|              NULL |		+ is a MySQL operator
| 4.055849718894897 |
|              NULL |
|              NULL |
+-------------------+
```

### Where do they live in the BI codebase?
Every scalar function we support has either 1 or 2 implementations in the codebase: an in-memory version, and possibly also a pushdown version. The in-memory implementations are in `expr_functions_scalar.go` (see `scalarFuncMap` for a list), and the pushdown implementations are cases under a long switch in `TranslateExpr` in `expr_translators.go`. 


### Why do we have two implementations for some scalar functions? What's the difference? What do you mean by pushdown?
A: To illustrate this let's choose a particular scalar function, `upper`. [Here it is in the MySQL docs](https://dev.mysql.com/doc/refman/5.7/en/string-functions.html#function_upper), but it works much as you'd expect: it converts a string to uppercase. 

In the in-memory implementation (see `ucaseFunc`) - where we have in-memory the data we're working with (in this case, the string) -  we simply use Golang's `strings.ToUpper` function to get the desired result. Easy! For example, I can type `select upper('hello')` into the MySQL shell and it will give me back HELLO. 

Now, imagine a case where we don't have the data loaded into memory when the function is called -- for example, if I wrote `select upper(b) from bar` (where `b` is a column in my table `bar`). 

One way we could solve this is to load the column `b` into memory by asking MongoDB for the data, and then use the in-memory implementation of `upper` on each row in the column. 

But… there's a better way to do this! Using the aggregation pipeline, we can ask Mongo to do the conversion to uppercase _before_ it returns the data to us. Then, rather than loading everything into memory and then processing it, we can just directly give the user the results. In the aggregation pipeline, you could accomplish this with something like: 
```javascript
stage1 = {
  $project: {
    'upper(b)': {
      $toUpper: '$b'		// if I have a collection bar, where  
    }						// each document has a field b...  
  }
};

db.bar.aggregate([stage1]);
```
This way, MongoDB can perform the conversion to uppercase as it fetches the data. This is much faster and more performant than executing in-memory, so we aim to push functions down to MongoDB whenever possible. 

But wait - if we can do this in MongoDB, why do we need the in-memory implementations at all?

Recall that scalar functions can be called with literal values -- if I type `select upper('hello')` in the MySQL shell, I'm not querying the database at all, so pushing down wouldn't make sense. 

Another reason - sometimes we are unable to push down a function at all due to lacking operators in the aggregation pipeline. For example, the MySQL `sin()` function cannot be pushed down, because there is no aggregation operator to take the sine of a number.

And a third - sometimes we just can't push a function down in certain cases/for certain inputs, and we need the in-memory version for those times. Some examples: 
* While we can push down `select log(b) from bar`, we cannot push down `select log(sin(b)) from bar`, because the inner sine part must be evaluated in-memory first before we calculate the log. 
* Some pushdown implementations use operators that only exist in newer versions of MongoDB, and if the user is not up-to-date then we have to evaluate in memory. 
* MongoDB has a limit on the depth of document nesting, if an expression is too complex and the pushdown would result in surpassing that depth, it cannot be pushed down.
* Generally, any case where we return `nil, false` in `TranslateExpr` is one of these cases we could not push down for some reason. 


### There are several methods (Evaluate, normalize…) associated with each scalar function type. What do they do?


Method Name | What it does | Does my scalar function need it?
--- | --- | ---
`Evaluate` | Evaluates the scalar function in-memory. | Always
`Validate` | Checks that the scalar function is being passed the correct number of arguments. | Always
`EvalType` | Returns what SQL type this scalar function returns. | Always
`Normalize` | Checks for cases where we can short-circuit evaluation because of some special input type, like a SQLNull.  | Only if there is an input type where we can short-circuit.
`Reconcile` | Converts function arguments to the desired input types. | If MySQL supports cases like those in the footnote (string instead of number, etc.) for the function, then yes.
`RequiresEvalCtx` | Indicates that an evaluation context is required to execute this function. | Only if you need the described purpose.
`FuncToAggregationLanguage` | Translates the referred scalar function to its equivalent MDB Aggregation Language | If possible.

(1) For example: try executing `select left('hi', 2)` and then `select left('hi', '2')`. Even though we use a string the second time, it still works, because `leftFunc`'s reconcile method converts the second argument to a SQL Int.

(2) For example, the `sleep` function implements this: the reason is that, if one calls `select sleep(1) from bar` where `bar` is a table with `x` rows, MySQL sleeps for `x` seconds. Thus, even though we don’t use the values in `bar`, we need an `EvalCtx` (containing the rows in `bar`) to sleep for the right amount of time. 


### What’s the best way to figure out how to push down a scalar function? 
My approach has been to create a bunch of integration test data, load it into Mongo by running the integration tests, and then work in the Mongo shell to slowly build out an aggregation stage to accomplish what I want and handle more cases. Once you get it working with Mongo directly, you just need to translate it into Go. 

Keep in mind that you may need to be very creative! The [Mongo aggregation operator docs](https://docs.mongodb.com/master/reference/operator/aggregation/) will be essential.

### How do I test them? 
Scalar functions should have both unit tests, and integration tests. You'll want to cover generally the same cases in both. 

#### Unit Tests
The unit tests live in `evaluator/expr_test.go`. Typically, you'll want to add cases under:
* `TestEvaluates`: these tests call the scalar functions with literal values. 
* `TestTranslateExpr`: these tests check that the aggregation pipeline translations of scalar function calls are what we expect them to be. 
* `TestExprNoPushdown`: Sometimes a scalar function should not be pushed down - either because we can't push it down at all, or because there are some special cases where we can't. These tests assert that the cases provided are indeed evaluated in memory and not pushed down. 

To run all of the evaluator unit tests:
```
cd evaluator
go test -v 
```
To only run one of the test functions, you can do:
```
go test -v -run TestEvaluates
``` 
etc. 

#### Internal Integration Tests
Most of the scalar function integration tests are in `testdata/suites/internal/functions.yml`.

To get started writing them, I would suggest you first simply run the existing ones. First, start sqlproxy with the test schema:
```
sqlproxy -vv --schema testdata/resources/internal
```
And then to load in the test data and run the tests:
```
go test -v -run /internal -automate data
```

After the tests run, you can connect to sqlproxy with the MySQL shell and browse the existing test data that way. If possible, try to write internal integration tests using that existing test data. If you need to test something that is impossible with the current data, you can add new test data by:
1) Adding the schema in `testdata/resources/schema/internal.yml`
2) Adding the data itself in `testdata/suites/internal/_suite.yml`
3) Using the `-automate data` flag the next time you run the tests 

Also useful - to run only tests with names matching `name`:
```
go test -v -run /internal/name
```

