#!/bin/bash
grep -E '^## ' Makefile | sed 's/## //' | column -t -s ':'
