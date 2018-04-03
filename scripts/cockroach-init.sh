#!/bin/bash
/cockroach/cockroach.sh start --insecure
/cockroach/cockroach.sh user set $COCKROACH_USER --insecure -u root
/cockroach/cockroach.sh sql -e "CREATE DATABASE $COCKROACH_DB;" --insecure -u root
/cockroach/cockroach.sh sql -e "GRANT ALL ON DATABASE recipes TO $COCKROACH_USER;" --insecure -u root
