#!/bin/bash
export JAVA_HOME=/usr/lib/jvm/java-16-openjdk/
echo "compiling"
CGO_CFLAGS="-I$JAVA_HOME/include -I$JAVA_HOME/include/linux" go build -buildmode=c-shared -o "$1"/libdsncore.so .
echo "done"