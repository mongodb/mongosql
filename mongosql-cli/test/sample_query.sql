SELECT customerAge, COUNT(*) FROM sample_supplies.sales GROUP BY customer.age AS customerAge limit 10;
