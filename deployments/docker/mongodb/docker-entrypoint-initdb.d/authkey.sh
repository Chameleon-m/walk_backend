#!/bin/bash -e
echo "START chmod /auth/file.key"
chmod 400 /auth/file.key
echo "END chmod /auth/file.key"
echo "START chown 999:999 /auth/file.key"
chown 999:999 /auth/file.key
echo "END chown 999:999 /auth/file.key"