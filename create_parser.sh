#!/usr/bin/env bash
mvn clean compile assembly:single -f ".\src\w2vtokenizer\pom.xml" || echo -e "\033[33mIs a java instance still running? Reset all notebooks!\033[0m"
