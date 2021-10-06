@echo off
echo compiling
set JAVA_HOME=C:\PROGRA~1\Java\jdk-17
set CGO_CFLAGS=-I%JAVA_HOME%\include -I%JAVA_HOME%\include\win32
go build -buildmode=c-shared -o %1/dsncore.dll .
echo done