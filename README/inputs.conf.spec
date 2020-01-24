[sqlinput://default]
*This is how the wattson app is configured

query = <value>
*The sql query to execute for the data to push towards Splunk.

connectionstring = <value>
*The connectionString to the database.

environment = <value>
*The application environment.

checkpoint = <value>
*Checkpoint enabled (if enabled, the query must contains the placeholded {{checkpoint}})
* ex: SELECT * FROM table id > {{checkpoint}}

checkpoint_id_query = <value>
*The query that retreive the id checkpoint (nb: To enable only when checkpoint field enabled)
* ex: SELECT MAX(id) FROM table

checkpoint_id_start = <value>
* The First Id (For the first run)