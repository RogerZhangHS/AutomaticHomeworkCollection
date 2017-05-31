#! /bin/sh
### BEGIN INIT INFO
# Provides:          webappStudent.sh
# Required-Start:    $remote_fs $syslog
# Required-Stop:     $remote_fs $syslog
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: Runs the client WebApp binary
# Description:       Makes sure the client WebApp binary starts on boot
### END INIT INFO

case "$1" in
  start)
    echo "Starting WebAppStudent"
    /home/pi/WebAppStudent -templates /home/pi/goworkspace/src/PiScanStudent/client/ui/templates >> /home/pi/WebAppStudent.log 2>&1
    ;;
  stop)
    echo "Stopping WebApp"
    killall WebAppStudent
    ;;
  *)
    echo "Usage: /etc/init.d/webappStudent.sh {start|stop}"
    exit 1
    ;;
esac

exit 0
