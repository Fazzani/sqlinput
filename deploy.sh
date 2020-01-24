#!/bin/bash

export SPLUNK_PATH=/Applications/Splunk

function deploy() {
    mkdir -p ./dest/sqlinput/{bin,README} && \
    mkdir -p $SPLUNK_PATH/etc/apps/sqlinput && \
    echo -e "\nBuiding sqlinput" && \
    go build -o ./dest/sqlinput/bin . && \
    {
        cp -rf ./README ./dest/sqlinput/
        cp -rf ./dest/sqlinput/ $SPLUNK_PATH/etc/apps/sqlinput
    }
}

function splunk_clear_index(){
    $SPLUNK_PATH/bin/splunk stop && \
    rm -rf $SPLUNK_PATH/var/lib/splunk/modinputs/sqlinput/* && \
    $SPLUNK_PATH/bin/splunk clean eventdata && \
    $SPLUNK_PATH/bin/splunk start
}

# /Applications/Splunk/bin/splunk restart
function test(){
    read -r -d '' xml << EOF
    <input>
        <server_host>localhost</server_host>
        <server_uri>https://127.0.0.1:8020</server_uri>
        <session_key>123102983109283019283</session_key>
        <checkpoint_dir>/opt/splunk/var/lib/splunk/sqlinput</checkpoint_dir>
        <configuration>
            <stanza name="sqlinputScheme:sqlinput">
                <param name="query">SELECT * FROM volumes_staging.expected_feeds WHERE id > {{.checkpoint}}</param>
                <param name="connectionstring">host='localhost' dbname='datatrust' user='dt_user' password='dtpass'</param>
                <param name="environment">dev</param>
                <param name="checkpoint">true</param>
                <param name="checkpoint_id_query">SELECT MAX(id) FROM volumes_staging.expected_feeds</param>
                <param name="checkpoint_id_start">0</param>
            </stanza>
        </configuration>
    </input>
EOF

    go build -o ./dest/sqlinput/bin && \
    echo "$xml" | ./dest/sqlinput/bin/sqlinput
}

if [ $# -lt 1 ]
then
    echo "Usage : $0 [test|deploy]"
    exit
fi


case "$1" in
    "test"|"-t")
        echo -n "Mode test"
        test
    ;;
    "deploy"|"-d")
        echo -n "SQL deployment on SPLUNK"
        deploy
    ;;
    "clear"|"-c")
        echo -n "Clear SPLUNK indices"
        splunk_clear_index
    ;;
    *)
    echo -n "unknown"
    ;;
esac