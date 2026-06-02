#!/system/bin/sh
MODDIR="$(cd "$(dirname "$0")" && pwd)"
WORK_DIR="/data/adb/sshcustom"

start() {
    sh "$WORK_DIR/scripts/ssh.service" start
}

stop() {
    sh "$WORK_DIR/scripts/ssh.service" stop
}

case "$1" in
    start) start ;;
    stop) stop ;;
esac
