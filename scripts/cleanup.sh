#!/bin/bash

# This script is used to delete expired files
# For this script to work, the uploads dir needs to be passed as an arg
# echo "/tmp/uploads" | bash cleanup.sh

set -e

echo "Checking for expired files..."

uploadDir=$1

if [[ uploadDir -eq "" ]];then
    echo $uploadDir
    echo "The uploads dir needs to be passed as arg"
    exit 1
fi

echo "Listing files in $uploadDir" 

ls $1