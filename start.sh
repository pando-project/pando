#!/usr/bin/env bash

if [ -d "$HOME/.pando" ]; then
  /opt/pando daemon
else
  /opt/pando init && /opt/pando daemon
fi