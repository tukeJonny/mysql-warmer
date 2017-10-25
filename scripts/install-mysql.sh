#!/usr/bin/env bash

echo "[1/3] Download deb package..."
wget http://dev.mysql.com/get/mysql-apt-config_0.6.0-1_all.deb \ 
    && sudo dpkg -i mysql-apt-config_0.6.0-1_all.deb \
    && rm -f mysql-apt-config_0.6.0-1_all.deb
echo "Done."

echo "[2/3] Install mysql-server..."
sudo apt update -y \
    && sudo apt install -y mysql-server
echo "Done."

echo "[3/3] Check MySQL version..."
mysql --version
echo "Done."

echo "[4/4] Initialize test db..."
mysql -uroot < seed.sql 
echo "Done."
