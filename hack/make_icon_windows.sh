#!/usr/bin/env bash

echo '// Code generated for package icon by 2goarray DO NOT EDIT.' > "${OUTPUT}"
{ echo '//go:build windows' ;echo '// +build windows' ;} >> "${OUTPUT}"
2goarray Data icon < "${INPUT}" >> "${OUTPUT}"
