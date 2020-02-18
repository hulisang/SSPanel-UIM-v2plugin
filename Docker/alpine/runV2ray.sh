#!/usr/bin/env bash

if [ ! -z "${api_port}" ]
    then
          sed -i "s|\"port\": 2333,|\"port\": ${api_port},|"  "/etc/v2ray/config.json"
fi
if [ ! -z "${sspanel_url}" ]
    then
         sed -i "s|\"https://google.com\"|\"${sspanel_url}\"|g" "/etc/v2ray/config.json"
fi
if [ ! -z "${key}" ]
    then
         sed -i "s/\"55fUxDGFzH3n\"/\"${key}\"/g" "/etc/v2ray/config.json"
fi

if [ ! -z "${node_id}" ]
    then
         sed -i "s/123456/${node_id}/g" "/etc/v2ray/config.json"
fi

if [ ! -z "${speedtest}" ]
    then
        sed -i "s/\"SpeedTestCheckRate\": 6/\"SpeedTestCheckRate\": ${speedtest}/g" "/etc/v2ray/config.json"
fi

if [ ! -z "${checkrate}" ]
    then
        sed -i "s/\"checkRate\": 60/\"checkRate\": ${checkrate}/g" "/etc/v2ray/config.json"
fi

if [ ! -z "${downWithPanel}" ]
    then
       sed -i "s/\"downWithPanel\": 1/\"downWithPanel\": ${downWithPanel}/g" "/etc/v2ray/config.json"
fi

if [ ! -z "${MYSQLHOST}" ]
    then
       sed -i "s|"https://bing.com"|"${MYSQLHOST}"|g" "/etc/v2ray/config.json"
fi

if [ ! -z "${MYSQLDBNAME}" ]
    then
       sed -i "s/"demo_dbname"/"${MYSQLDBNAME}"/g" "/etc/v2ray/config.json"
fi

if [ ! -z "${MYSQLUSR}" ]
    then
       sed -i "s|\"demo_user\"|\"${MYSQLUSR}\"|g" "/etc/v2ray/config.json"
fi
if [ ! -z "${MYSQLPASSWD}" ]
    then
      sed -i "s/"demo_dbpassword"/"${MYSQLPASSWD}"/g" "/etc/v2ray/config.json"
fi
if [ ! -z "${MYSQLPORT}" ]
    then
      sed -i "s/3306,/${MYSQLPORT},/g" "/etc/v2ray/config.json"
fi

if [ ! -z "${PANELTYPE}" ]
    then
      sed -i "s|\"paneltype\": 0|\"paneltype\": ${PANELTYPE}|g" "/etc/v2ray/config.json"
fi

if [ ! -z "${usemysql}" ]
    then
      sed -i "s|\"usemysql\": 0|\"usemysql\": ${usemysql}|g" "/etc/v2ray/config.json"
fi

if [ ! -z "${LDNS}" ]
    then
      sed -i "s|\"localhost\"|\"${LDNS}\"|g" "/etc/v2ray/config.json"
fi
if [ ! -z "${CF_Key}" ]
then
  sed -i "s|\"bbbbbbbbbbbbbbbbbb\"|\"${CF_Key}\"|g" "/etc/v2ray/config.json"
fi
if [ ! -z "${CF_Email}" ]
then
  sed -i "s|\"rico93@outlxxxxxxxxxx.com\"|\"${CF_Email}\"|g" "/etc/v2ray/config.json"

fi


if [ ! -z "${NodeUserLimited}" ]
    then
        sed -i "s/\"NodeUserLimited\": 4/\"NodeUserLimited\": ${NodeUserLimited}/g" "/etc/v2ray/config.json"
fi

if [ ! -z "${UseIP}" ]
then
  sed -i "s|\"UseIP\"|\"${UseIP}\"|g" "/etc/v2ray/config.json"

fi

cat /etc/v2ray/config.json
v2ray -config=/etc/v2ray/config.json
