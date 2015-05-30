#!/usr/bin/env 

docker create --volumes=$(pwd):/jot --name jot-data ubuntu

docker run --volumes-from jot-data ubuntu ./jot/jotserver 