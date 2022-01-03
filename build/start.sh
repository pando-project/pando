#!/usr/bin/env bash

if [ -d "$HOME/.pando" ]; then
  /opt/pando-server daemon && /opt/go-swagger serve -F=swagger /opt/swagger.yml -p 5000
else
  /opt/pando-server init && /opt/pando daemon && /opt/go-swagger serve -F=swagger /opt/swagger.yml -p 5000
fi