#!/bin/bash

mkdir deploy
go build
mv hypermedia deploy/
cp -r static/ deploy/
cp -r tmpl/ deploy/

scp -r deploy/. craftingTable:/home/matej/hypermedia
ssh root@craftingTable systemctl restart hypermedia.service

rm -rf deploy/
