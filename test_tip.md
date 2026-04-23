# Window functions do not collapse rows

Unlike `GROUP BY`, the `OVER()` clause lets you compute aggregates while retaining every individual row. 
This makes window functions the right tool for running totals, rankings, and lead/lag comparisons - 
tasks where `GROUP BY` would discard the detail you need.
