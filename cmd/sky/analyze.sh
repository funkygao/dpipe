#!/bin/sh
#=================================
# rs_user
#=================================

curl -X GET http://localhost:8585/tables/rs_user

curl -X GET http://localhost:8585/tables/rs_user/stats

curl -X POST http://localhost:8585/tables/rs_click/query -d '{
    "steps": [
        {
            "type": "selection",
            "name": "0",
            "dimensions": ["ver", "action"],
            "fields":[
                {
                    "name":"count",
                    "expression":"count()"
                }
            ]
        }
    ]
}'

curl -s -XPOST localhost:8585/tables/rs_click/query -d '{
    "sessionIdleTime": 7200,
    "steps": [
        {
            "type": "condition", 
            "expression": "true",
            "within": [0, 0],
            "steps": [
                {
                    "type": "selection",
                    "name": "0",
                    "dimensions": ["action"],
                    "fields": [
                        {
                            "name": "count",
                            "expression": "count()"
                        }
                    ]
                }
            ]
        }
    ]
}'

