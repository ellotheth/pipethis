#!/bin/bash

# This should never be done. 

curl https://github.com/ellotheth/pipethis/releases/download/v0.1/pipethis-0.1-linux-amd64.tar.bz2 | bunzip -d > /usr/bin/pipethis
chmod 755 /usr/bin/pipethis

