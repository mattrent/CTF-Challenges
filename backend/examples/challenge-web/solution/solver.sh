#!/bin/bash

curl http://web.ctf | grep "flag{web-example}"

if [ $? -eq 0 ]
then
    echo "Test: Succeeded"
    exit 0
else
    echo "Test: Failed"
    exit 1
fi
