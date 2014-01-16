#!/bin/sh
#=================================
# rs_user
#=================================

curl -X GET http://localhost:8585/tables/rs_user

curl -X GET http://localhost:8585/tables/rs_user/stats

curl -X POST http://localhost:8585/tables/rs_user/query -d '{
    "steps": [
        {"type":"visit","fields":[{"name":"count","expression":"count()"}]}
    ]
}'
