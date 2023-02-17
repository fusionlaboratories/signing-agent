#!/bin/bash

ACTION_ID=$1

curl -X 'DELETE' "http://localhost:8007/api/v1/client/action/$ACTION_ID"
