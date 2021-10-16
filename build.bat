@echo off
echo compiling
set JAVA_HOME=C:\PROGRA~1\Java\jdk-17
set CGO_CFLAGS=-I%JAVA_HOME%\include -I%JAVA_HOME%\include\win32
go build .
dsn-go.exe