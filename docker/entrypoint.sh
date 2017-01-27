#!/bin/bash
/easyssh -e="$easyssh_executor" -f="$easyssh_filter" -d="$easyssh_discoverer" "$@"
