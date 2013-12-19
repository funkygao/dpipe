curl localhost:9200/_aliases?pretty

curl -XPOST localhost:9200/_aliases -d '{
    "actions": [{
        "remove": {
            "index": "rs_2013_12",
            "alias": "rs_latest"
        },
        "add": {
            "index": "fun_rs_2013_12",
            "alias": "rs_latest"
        }
    }]
}'

curl -XPOST localhost:9200/_aliases -d '{
    "actions": [{
        "add": {
            "index": "fun_ffs_2013_12",
            "alias": "ffs_latest"
        }
    }]
}'
