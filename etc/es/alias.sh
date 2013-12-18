curl localhost:9200/_aliases?pretty

curl -XPOST localhost:9200/_aliases -d '{
    "actions": [{
        "add": {
            "index": "rs_2013_12",
            "alias": "rs_latest"
        }
    }]
}'
