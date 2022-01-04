#!/usr/bin/env bash

if [ -d "$HOME/.pando" ]; then
  /opt/pando-server daemon
else
  /opt/pando-server init && /opt/pando-server daemon
fi