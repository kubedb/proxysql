#!/bin/bash
set -e

## ProxySQL entrypoint
## ===================
##
## Supported environment variable:
##
## MONITOR_CONFIG_CHANGE={true|false}
## - Monitor /etc/proxysql.cnf for any changes and reload ProxySQL automatically

# use the current scrip name while putting log
script_name=${0##*/}

function timestamp() {
  date +"%Y/%m/%d %T"
}

function log() {
  local log_type="$1"
  local msg="$2"
  echo "$(timestamp) [$script_name] [$log_type] $msg"
}

# If command has arguments, prepend proxysql
if [ "${1:0:1}" = '-' ]; then
  CMDARG="$@"
fi

if [ $MONITOR_CONFIG_CHANGE ]; then

  log "INFO" "Env MONITOR_CONFIG_CHANGE=true"
  CONFIG=/etc/proxysql.cnf
  oldcksum=$(cksum ${CONFIG})

  # Start ProxySQL in the background
  proxysql --reload -f $CMDARG &

  log "INFO" "Configuring proxysql.."
  /usr/bin/configure-proxysql.sh

  log "INFO" "Monitoring $CONFIG for changes.."
  inotifywait -e modify,move,create,delete -m --timefmt '%d/%m/%y %H:%M' --format '%T' ${CONFIG} |
    while read date time; do
      newcksum=$(cksum ${CONFIG})
      if [ "$newcksum" != "$oldcksum" ]; then
        echo "++++++++++++++++++++++++++++++++++++++++++++++++++++++"
        echo "At ${time} on ${date}, ${CONFIG} update detected."
        echo "++++++++++++++++++++++++++++++++++++++++++++++++++++++"
        oldcksum=$newcksum
        log "INFO" "Reloading ProxySQL.."
        killall -15 proxysql
        proxysql --initial --reload -f $CMDARG
      fi
    done
fi

# Start ProxySQL with PID 1
exec proxysql -f $CMDARG &
pid=$!

log "INFO" "Configuring proxysql.."
/usr/bin/configure-proxysql.sh

log "INFO" "Waiting for proxysql ..."
wait $pid
