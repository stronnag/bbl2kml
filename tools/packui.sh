#!/usr/bin/env bash

echo -en 'package main
func GetUIString() string {
	return `'
xmllint --noblanks $1
echo -e '`
}'
