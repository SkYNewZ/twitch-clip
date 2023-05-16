#!/usr/bin/env bash

echo '// Code generated for package icon by 2goarray DO NOT EDIT.' > "${OUTPUT}"
{ echo '//go:build !windows' ;} >> "${OUTPUT}"
go run github.com/cratonica/2goarray@latest Data icon < "${INPUT}" >> "${OUTPUT}"
