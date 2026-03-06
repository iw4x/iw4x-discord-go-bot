#!/bin/sh
# for any future readers, this script has been kept strictly POSIX and shellcheck compliant
# if it appears to be written moderately strangely

# modify as needed
logfile="${PWD}/chatlog.json"

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

./logreader.sh -m <message_id> -c <channel_id> ... -e -d -t

${BOLD}Options:${CLEAR}
    -m  Specify the message ID to query the log for

    -c  Specify the channel ID to query the log for

    -a  Specify the author ID to query the log for

    -u  Specify the author username to query the log for

    -n  Specify the author nickname to query the log for

    -d  Search for only deleted messages (does not take a value)

    -e  Search for only edited messages (does not take a value)

    -t  Search for only attachments (does not take a value)"

    exit
}

case "$1" in
    help|--h|-h|'') help ;;
    *)
        while getopts "m:c:a:u:n:det" opts ; do
            case "${opts}" in
                m) message_id="$OPTARG" ;;
                c) channel_id="$OPTARG" ;;
                a) author_id="$OPTARG";;
                u) author_username="$OPTARG" ;;
                n) author_nickname="$OPTARG" ;;
                d) deleted=true ;;
                e) edited=true ;;
                t) attachment=true ;;
                *) help ;;
            esac
        done
    ;;
esac

# this is going to have to be slightly messy given the amount of ways this can be invoked, my apologies
query='.[]'

[ -n "$message_id" ] &&
    query="$query | select(.message_ID == \"$message_id\")"

[ -n "$channel_id" ] &&
    query="$query | select(.channel_ID == \"$channel_id\")"

[ -n "$author_id" ] &&
    query="$query | select(.author_ID == \"$author_id\")"

[ -n "$author_username" ] &&
    query="$query | select(.author_username == \"$author_username\")"

[ -n "$author_nickname" ] &&
    query="$query | select(.author_nickname == \"$author_nickname\")"

[ -n "$deleted" ] &&
    query="$query | select(.type == \"deletion\")"

[ -n "$edited" ] &&
    query="$query | select(.type == \"edit\")"

[ -n "$attachment" ] &&
    query="$query | select(.attachments | length > 0)"

jq "$query" "$logfile" || die "Query invalid."
