-- Per-account profit rollup, joined to each customer's unwound accounts.
--
-- Two equivalent forms are shown: derived tables first, then CTEs. They are
-- independent statements -- run them one at a time, not as a single script.
--
-- Notes:
--   * `total` is typed as a string in the schema, so it must be CAST to DOUBLE.
--   * The CASE has no ELSE. That is safe here because every transaction_code in
--     sample_analytics.transactions is 'buy' or 'sell'. An unrecognized code
--     would otherwise make the REDUCE accumulator NULL for the rest of the
--     array, so add ELSE `value` if the data can carry other codes.
--   * Derived-table form only: the right side of a JOIN must be a derived
--     table, not a bare SELECT.
--   * CTE form only: `WITH <name> AS (<query>)[, ...] (<select>)` -- no
--     trailing comma, and the final SELECT must be parenthesized.

-- Form 1: derived tables.
SELECT my_custom_query.username,
       my_custom_query.user_account,
       user_account_profits.profit
FROM (
  SELECT username, accounts AS user_account
  FROM UNWIND(sample_analytics.customers WITH PATH => accounts, INDEX => idx)
) AS my_custom_query
LEFT JOIN (
  SELECT account_id,
    REDUCE(transactions, 0.0,
      CASE
        WHEN this.transaction_code = 'buy'  THEN `value` - CAST(this.total AS DOUBLE)
        WHEN this.transaction_code = 'sell' THEN `value` + CAST(this.total AS DOUBLE)
      END
    ) AS profit
  FROM sample_analytics.transactions
) AS user_account_profits
ON my_custom_query.user_account = user_account_profits.account_id;

-- Form 2: CTEs. Identical to form 1, but as CTES
WITH my_custom_query AS (
    SELECT username, accounts AS user_account
    FROM UNWIND(sample_analytics.customers WITH PATH => accounts, INDEX => idx)
  ),
  user_account_profits AS (
    SELECT account_id,
      REDUCE(transactions, 0.0,
        CASE
          WHEN this.transaction_code = 'buy'  THEN `value` - CAST(this.total AS DOUBLE)
          WHEN this.transaction_code = 'sell' THEN `value` + CAST(this.total AS DOUBLE)
        END
      ) AS profit
    FROM sample_analytics.transactions
  )
(SELECT my_custom_query.username,
        my_custom_query.user_account,
        user_account_profits.profit
 FROM my_custom_query
 LEFT JOIN user_account_profits
   ON my_custom_query.user_account = user_account_profits.account_id);
