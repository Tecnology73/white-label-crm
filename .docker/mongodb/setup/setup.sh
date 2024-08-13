#!/usr/bin/env bash

if [ ! -f /data/init.flag ]; then
  mongosh mongodb://crm-mongo:27017 setup.js
  touch /data/init.flag
fi
