## Connection pool in database 

Instead of creating an new connection for each client trying to connection to the database (which is an expensive operation), the database can cache database connection which is called the connection pool. The connection pool contains a number of pre-database connections that can be given to a new user for each client request. This makes sure that a new connection is not created only when a client requests for the database connection.


## Placeholders in the SQL query 

MySQL supports placeholder using the ? notation
PostgresSQL supports placeholder using the $N notation, something like `INSERT INTO <table_name> values ($1, $2..., $N)`

Why do we need placeholders?
Prevents SQL injection

How does this work?
It creates a prepared statement in the database, the database parses and compiles the statement and stores it ready for execution.
When values are given to the database then the values are placed according to their order in the placeholder and the query is executed. Since the parameters are transmitted later, after the statement has been compiled, the database treats the values as pure values and they do not change the intent of the query

## The returning clause in postgres
