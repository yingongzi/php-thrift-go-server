#!/usr/bin/env bash
thrift -gen go php-go.thrift
thrift -gen php php-go.thrift
