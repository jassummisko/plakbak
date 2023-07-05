#!/bin/bash
go build -o bin/plakbak -ldflags="-s -w" &&
upx --best --lzma bin/plakbak &&
tar zcvf releases/plakbak.tar.gz bin/