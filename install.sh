#!/usr/bin/bash
go build
mv webby /usr/bin/
cp webby.service /etc/systemd/system/