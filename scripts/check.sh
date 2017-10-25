#!/usr/bin/env bash

check_mem_usage() {
    TOTAL_MEM_USAGE=$(free | awk '/Mem/ {print $2;}')
    FREE_MEM_USAGE=$(free  | awk '/buffers\/cache/ {print $4;}')
    result=$((100-100*${FREE_MEM_USAGE}/${TOTAL_MEM_USAGE}))
}

BEFORE=`check_mem_usage`
./mysql-warmup
AFTER=`check_mem_usage`

if [ $AFTER -gt $BEFORE ]; then
    echo "[+] MySQL InnoDB BufferPool used."
    exit 0
else
    echo "[-] MySQL InnoDB BufferPool not used."
    exit 1
fi

