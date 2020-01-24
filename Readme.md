# SQL modular input for SPLUNK

This module allows the injection of data from any **Postgres** database to Splunk instance. Through a simple request and a connectionString database.
> this module supports **checkpoints**.

## DEV

### Config input example

```xml
<input>
  <server_host>myHost</server_host>
  <server_uri>https://127.0.0.1:8020</server_uri>
  <session_key>123102983109283019283</session_key>
  <checkpoint_dir>/opt/splunk/var/lib/splunk/sqlinput</checkpoint_dir>
  <configuration>
    <stanza name="myScheme://aaa">
        <param name="query">SELECT * FROM table WHERE id > {{.checkpoint}}</param>
        <param name="connectionstring">host='localhost' dbname='db' user='user' password='pass'</param>
        <param name="environment">dev</param>
        <param name="checkpoint">true</param>
        <param name="checkpoint_id_query">SELECT MAX(id) FROM table</param>
        <param name="checkpoint_id_start">0</param>
    </stanza>
  </configuration>
</input>
```

### Commands

```shell
export SPLUNK_PATH=/Applications/Splunk

# For local testing
./deploy -t

# for deployment
./deploy -d

# for restart Splunk server
$SPLUNK_PATH/bin/splunk restart
```

## SPL commands

```splunk
# last flow integrations as expected
index="sql_input" source="sqlinput://rec2_expected_feeds" 
| eval str_last_run="-1" .lower(frequency) 
| eval str_last_run=replace(str_last_run, "m", "mon") 
| eval expected_last_integ=relative_time(now(), str_last_run) 
| join type=left feed_name 
    [ search index="sql_input" source="sqlinput://rec2_feed_integration" 
    | stats latest(integration_end_datetime) as integration_end_datetime by feed_name 
    | eval int_end_dt=strptime(integration_end_datetime,"'%Y-%m-%d %H:%M:%S.%6Q %z %Z'") ] 
| where int_end_dt >= expected_last_integ 
| stats latest(integration_end_datetime) as integration_end_datetime by feed_name
```

- [SPL doc](https://docs.splunk.com/Documentation/Splunk/8.0.1/SearchReference/Stats)
- [Config files doc](https://docs.splunk.com/Documentation/Splunk/8.0.1/Admin/Appconf)

## TODO

- [ ] alerting