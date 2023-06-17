#!/usr/bin/bash
go build
mv webby /usr/bin/
cp webby.service /etc/systemd/system/
mkdir -p /srv/webby/website
