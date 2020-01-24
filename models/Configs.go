package models

import (
	"encoding/xml"
	"fmt"
	"os"
)

// ConfigInput Splunk modular input configuration (provided by splunk on stdin)
type ConfigInput struct {
	XMLName       xml.Name `xml:"input"`
	Text          string   `xml:",chardata"`
	ServerHost    string   `xml:"server_host"`
	ServerURI     string   `xml:"server_uri"`
	SessionKey    string   `xml:"session_key"`
	CheckpointDir string   `xml:"checkpoint_dir"`
	Configuration struct {
		Text   string `xml:",chardata"`
		Stanza struct {
			Text  string `xml:",chardata"`
			Name  string `xml:"name,attr"`
			Param []struct {
				Text string `xml:",chardata"`
				Name string `xml:"name,attr"`
			} `xml:"param"`
		} `xml:"stanza"`
	} `xml:"configuration"`
}

func (c ConfigInput) String() string {
	return fmt.Sprintf("Text: %s ServerHost: %s Configuration.Name: %s", c.ServerURI, c.ServerHost, c.Configuration.Stanza.Name)
}

// SplunkConfig Splunk configuration
type SplunkConfig interface {
	Get() error
}

// Get splunk configuration
func (conf *ConfigInput) Get() error {
	dec := xml.NewDecoder(os.Stdin)
	dec.Strict = false
	return dec.Decode(conf)
}

// SQLInputConfig Custom config input
type SQLInputConfig struct {
	Query             string
	ConnString        string
	Environment       string
	Checkpoint        bool
	CheckpointIDQuery string
	CheckpointIDStart string
}

// SCHEME Spluck modular input scheme
const SCHEME = `<scheme>
    <title>Postgres SQL input</title>
    <description>Postgres SQL pulling.</description>
    <use_external_validation>true</use_external_validation>
    <streaming_mode>simple</streaming_mode>
    <endpoint>
        <args>
            <arg name="query">
                <title>SQL Query</title>
                <description>The sql query to execute for the data to push towards Splunk.</description>
                <validation>
                    validate(match(lower('query'), '^select'), "SQL query must begin with SELECT")
                </validation>
                <data_type>string</data_type>
                <required_on_edit>true</required_on_edit>
                <required_on_create>true</required_on_create>
            </arg>

            <arg name="connectionstring">
                <title>The connection string</title>
                <description>The connectionString to the database.</description>
            </arg>

            <arg name="ENVIRONMENT">
                <title>environment</title>
                <description>The execution environment.</description>
            </arg>

            <arg name="checkpoint">
                <title>Checkpopint</title>
                <data_type>boolean</data_type>
                <validation>is_bool('checkpoint')</validation>
                <required_on_edit>false</required_on_edit>
                <required_on_create>false</required_on_create>
                <description>Enable checkpoint. (if enabled, the query must contains the placeholded {{.checkpoint}}, ex: SELECT * FROM table id > {{.checkpoint}}</description>
            </arg>

            <arg name="checkpoint_id_query">
                <title>Checkpoint Id query</title>
                <description>The query that retreive the id checkpoint (nb: To enable only when checkpoint field enabled) ex: SELECT MAX(id) FROM table</description>
                <required_on_edit>false</required_on_edit>
                <required_on_create>false</required_on_create>
			</arg>

			<arg name="checkpoint_id_start">
                <title>Checkpoint Id start</title>
                <description>The First Id (For the first run)</description>
                <required_on_edit>false</required_on_edit>
                <required_on_create>false</required_on_create>
            </arg>

        </args>
    </endpoint>
</scheme>`
