#!/bin/sh
# for any future readers, this script has been kept strictly POSIX and shellcheck compliant
# if it appears to be written moderately strangely

# modify as needed
logfile="${PWD}/iw4xchat.log"

# probably don't modify
RED='\033[0;31m'
BOLD='\033[1m'
CLEAR='\033[0m'

die() {
    printf "%b\n" "${RED}FATAL:${CLEAR} $1"
    exit 1
}

# jq is required to read the log with this shell script
# everything else here should be covered by coreutils
! command -v jq > /dev/null &&
    die "jq binary not found in PATH, cannot continue"

help() {
    printf "%b\n\n" "${BOLD}Usage:${CLEAR}

./logreader.sh -a <author_id> -c <channel_id> -m <message_id> -e -d

${BOLD}Options:${CLEAR}
    -a  Specify the author ID to query the log for

    -c  Specify the channel ID to query the log for

    -m  Specify the message ID to query the log for

    -e  Search for only edited messages

    -d  Search for only deleted messages"

    exit
}

case "$1" in
    help|--h|-h|'') help ;;
    *)
        while getopts "a:c:m:ed" opts ; do
            case "${opts}" in
                a) author_id="$OPTARG";;
                c) channel_id="$OPTARG" ;;
                m) message_id="$OPTARG" ;;
                e) edited=true ;;
                d) deleted=true ;;
                *) help ;;
            esac
        done
    ;;
esac

# this is going to have to be slightly messy given the amount of ways this can be invoked, my apologies
query='.'

[ -n "$author_id" ] &&
    query="$query | select(.author_ID == \"$author_id\")"

[ -n "$channel_id" ] &&
    query="$query | select(.channel_ID == \"$channel_id\")"

[ -n "$message_id" ] &&
    query="$query | select(.message_ID == \"$message_id\")"

[ -n "$edited" ] &&
    query="$query | select(.type == \"edit\")"

[ -n "$deleted" ] &&
    query="$query | select(.type == \"deletion\")"

jq "$query" "$logfile" || die "Query invalid."
