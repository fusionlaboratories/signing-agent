#!/bin/bash

ACTION_ID=$1

curl -X 'PUT' "http://localhost:8007/api/v1/client/action/$ACTION_ID"
