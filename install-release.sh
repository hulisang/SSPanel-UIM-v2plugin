#!/bin/bash

# This file is accessible as https://install.direct/go.sh
# Original source is located at github.com/v2ray/v2ray-core/release/install-release.sh

# If not specify, default meaning of return value:
# 0: Success
# 1: System error
# 2: Application error
# 3: Network error

CUR_VER=""
NEW_VER=""
ARCH=""
VDIS="64"
ZIPFILE="/tmp/v2ray/v2ray.zip"
V2RAY_RUNNING=0
VSRC_ROOT="/tmp/v2ray"
EXTRACT_ONLY=0
ERROR_IF_UPTODATE=0

CMD_INSTALL=""
CMD_UPDATE=""
SOFTWARE_UPDATED=0

SYSTEMCTL_CMD=$(command -v systemctl 2>/dev/null)
SERVICE_CMD=$(command -v service 2>/dev/null)

CHECK=""
FORCE=""
HELP=""

#######color code########
RED="31m"      # Error message
GREEN="32m"    # Success message
YELLOW="33m"   # Warning message
BLUE="36m"     # Info message


#########################
while [[ $# > 0 ]];do
    key="$1"
    case $key in
        -p|--proxy)
        PROXY="-x ${2}"
        shift # past argument
        ;;
        -h|--help)
        HELP="1"
        ;;
        -f|--force)
        FORCE="1"
        ;;
        -c|--check)
        CHECK="1"
        ;;
        --remove)
        REMOVE="1"
        ;;
        --version)
        VERSION="$2"
        shift
        ;;
        --extract)
        VSRC_ROOT="$2"
        shift
        ;;
        --extractonly)
        EXTRACT_ONLY="1"
        ;;
        -l|--local)
        LOCAL="$2"
        LOCAL_INSTALL="1"
        shift
        ;;
        --errifuptodate)
        ERROR_IF_UPTODATE="1"
        ;;
        --panelurl)
        PANELURL="$2"
        ;;
        --panelkey)
        PANELKEY="$2"
        ;;
        --nodeid)
        NODEID="$2"
        ;;
        --downwithpanel)
        DOWNWITHPANEL="$2"
        ;;
        --mysqlhost)
        MYSQLHOST="$2"
        ;;
        --mysqldbname)
        MYSQLDBNAME="$2"
        ;;
        --mysqluser)
        MYSQLUSR="$2"
        ;;
        --mysqlpasswd)
        MYSQLPASSWD="$2"
        ;;
        --mysqlport)
        MYSQLPORT="$2"
        ;;
        --speedtestrate)
        SPEEDTESTRATE="$2"
        ;;
        --paneltype)
        PANELTYPE="$2"
        ;;
        --usemysql)
        USEMYSQL="$2"
        ;;
        --ldns)
        LDNS="$2"
        ;;
        --cfkey)
        CFKEY="$2"
        ;;
        --cfemail)
        CFEMAIL="$2"
        ;;
        --nodeuserlimited)
        NODEUSERLIMITED="$2"
        ;;
        --useip)
        USEIP="$2"
        ;;
        *)
                # unknown option
        ;;
    esac
    shift # past argument or value
done

###############################
colorEcho(){
    COLOR=$1
    echo -e "\033[${COLOR}${@:2}\033[0m"
}

sysArch(){
    ARCH=$(uname -m)
    if [[ "$ARCH" == "i686" ]] || [[ "$ARCH" == "i386" ]]; then
        VDIS="32"
    elif [[ "$ARCH" == *"armv7"* ]] || [[ "$ARCH" == "armv6l" ]]; then
        VDIS="arm"
    elif [[ "$ARCH" == *"armv8"* ]] || [[ "$ARCH" == "aarch64" ]]; then
        VDIS="arm64"
    elif [[ "$ARCH" == *"mips64le"* ]]; then
        VDIS="mips64le"
    elif [[ "$ARCH" == *"mips64"* ]]; then
        VDIS="mips64"
    elif [[ "$ARCH" == *"mipsle"* ]]; then
        VDIS="mipsle"
    elif [[ "$ARCH" == *"mips"* ]]; then
        VDIS="mips"
    elif [[ "$ARCH" == *"s390x"* ]]; then
        VDIS="s390x"
    elif [[ "$ARCH" == "ppc64le" ]]; then
        VDIS="ppc64le"
    elif [[ "$ARCH" == "ppc64" ]]; then
        VDIS="ppc64"
    fi
    return 0
}

downloadV2Ray(){
    rm -rf /tmp/v2ray
    mkdir -p /tmp/v2ray
    colorEcho ${BLUE} "Downloading V2Ray."
    DOWNLOAD_LINK="https://github.com/ns-sp4ce/V2Ray-With-SSpanel/releases/download/${NEW_VER}/v2ray-linux-${VDIS}.zip"
    curl ${PROXY} -L -H "Cache-Control: no-cache" -o ${ZIPFILE} ${DOWNLOAD_LINK}
    if [ $? != 0 ];then
        colorEcho ${RED} "Failed to download! Please check your network or try again."
        return 3
    fi
    return 0
}

installSoftware(){
    COMPONENT=$1
    if [[ -n `command -v $COMPONENT` ]]; then
        return 0
    fi

    getPMT
    if [[ $? -eq 1 ]]; then
        colorEcho ${RED} "The system package manager tool isn't APT or YUM, please install ${COMPONENT} manually."
        return 1
    fi
    if [[ $SOFTWARE_UPDATED -eq 0 ]]; then
        colorEcho ${BLUE} "Updating software repo"
        $CMD_UPDATE
        SOFTWARE_UPDATED=1
    fi

    colorEcho ${BLUE} "Installing ${COMPONENT}"
    $CMD_INSTALL $COMPONENT
    if [[ $? -ne 0 ]]; then
        colorEcho ${RED} "Failed to install ${COMPONENT}. Please install it manually."
        return 1
    fi
    return 0
}

# return 1: not apt, yum, or zypper
getPMT(){
    if [[ -n `command -v apt-get` ]];then
        CMD_INSTALL="apt-get -y -qq install"
        CMD_UPDATE="apt-get -qq update"
    elif [[ -n `command -v yum` ]]; then
        CMD_INSTALL="yum -y -q install"
        CMD_UPDATE="yum -q makecache"
    elif [[ -n `command -v zypper` ]]; then
        CMD_INSTALL="zypper -y install"
        CMD_UPDATE="zypper ref"
    else
        return 1
    fi
    return 0
}

extract(){
    colorEcho ${BLUE}"Extracting V2Ray package to /tmp/v2ray."
    mkdir -p /tmp/v2ray
    unzip $1 -d ${VSRC_ROOT}
    if [[ $? -ne 0 ]]; then
        colorEcho ${RED} "Failed to extract V2Ray."
        return 2
    fi
    if [[ -d "/tmp/v2ray/v2ray-${NEW_VER}-linux-${VDIS}" ]]; then
      VSRC_ROOT="/tmp/v2ray/v2ray-${NEW_VER}-linux-${VDIS}"
    fi
    return 0
}


# 1: new V2Ray. 0: no. 2: not installed. 3: check failed. 4: don't check.
getVersion(){
    if [[ -n "$VERSION" ]]; then
        NEW_VER="$VERSION"
        if [[ ${NEW_VER} != v* ]]; then
          NEW_VER=v${NEW_VER}
        fi
        return 4
    else
        VER=`/usr/bin/v2ray/v2ray -version 2>/dev/null`
        RETVAL="$?"
        CUR_VER=`echo $VER | head -n 1 | cut -d " " -f2`
        if [[ ${CUR_VER} != v* ]]; then
            CUR_VER=v${CUR_VER}
        fi
        TAG_URL="https://api.github.com/repos/ns-sp4ce/V2Ray-With-SSpanel/releases/latest"
        NEW_VER=`curl ${PROXY} -s ${TAG_URL} --connect-timeout 10| grep 'tag_name' | cut -d\" -f4`
        if [[ ${NEW_VER} != v* ]]; then
          NEW_VER=v${NEW_VER}
        fi
        if [[ $? -ne 0 ]] || [[ $NEW_VER == "" ]]; then
            colorEcho ${RED} "Failed to fetch release information. Please check your network or try again."
            return 3
        elif [[ $RETVAL -ne 0 ]];then
            return 2
        elif [[ `echo $NEW_VER | cut -d. -f-2` != `echo $CUR_VER | cut -d. -f-2` ]];then
            return 1
        fi
        return 0
    fi
}

stopV2ray(){
    colorEcho ${BLUE} "Shutting down V2Ray service."
    if [[ -n "${SYSTEMCTL_CMD}" ]] || [[ -f "/lib/systemd/system/v2ray.service" ]] || [[ -f "/etc/systemd/system/v2ray.service" ]]; then
        ${SYSTEMCTL_CMD} stop v2ray
    elif [[ -n "${SERVICE_CMD}" ]] || [[ -f "/etc/init.d/v2ray" ]]; then
        ${SERVICE_CMD} v2ray stop
    fi
    if [[ $? -ne 0 ]]; then
        colorEcho ${YELLOW} "Failed to shutdown V2Ray service."
        return 2
    fi
    return 0
}

startV2ray(){
    if [ -n "${SYSTEMCTL_CMD}" ] && [ -f "/lib/systemd/system/v2ray.service" ]; then
        ${SYSTEMCTL_CMD} start v2ray
    elif [ -n "${SYSTEMCTL_CMD}" ] && [ -f "/etc/systemd/system/v2ray.service" ]; then
        ${SYSTEMCTL_CMD} start v2ray
    elif [ -n "${SERVICE_CMD}" ] && [ -f "/etc/init.d/v2ray" ]; then
        ${SERVICE_CMD} v2ray start
    fi
    if [[ $? -ne 0 ]]; then
        colorEcho ${YELLOW} "Failed to start V2Ray service."
        return 2
    fi
    return 0
}

copyFile() {
    NAME=$1
    ERROR=`cp "${VSRC_ROOT}/${NAME}" "/usr/bin/v2ray/${NAME}" 2>&1`
    if [[ $? -ne 0 ]]; then
        colorEcho ${YELLOW} "${ERROR}"
        return 1
    fi
    return 0
}

makeExecutable() {
    chmod +x "/usr/bin/v2ray/$1"
}

installV2Ray(){
    # Install V2Ray binary to /usr/bin/v2ray
    mkdir -p /usr/bin/v2ray
    copyFile v2ray
    if [[ $? -ne 0 ]]; then
        colorEcho ${RED} "Failed to copy V2Ray binary and resources."
        return 1
    fi
    makeExecutable v2ray
    copyFile v2ctl && makeExecutable v2ctl
    copyFile geoip.dat
    copyFile geosite.dat
    colorEcho ${BLUE} "Setting TimeZone to Shanghai"
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
    date -s "$(curl -sI g.cn | grep Date | cut -d' ' -f3-6)Z"

    # Install V2Ray server config to /etc/v2ray
    if [[ ! -f "/etc/v2ray/config.json" ]]; then
        mkdir -p /etc/v2ray
        mkdir -p /var/log/v2ray
        cp "${VSRC_ROOT}/vpoint_vmess_freedom.json" "/etc/v2ray/config.json"
        if [[ $? -ne 0 ]]; then
            colorEcho ${YELLOW} "Failed to create V2Ray configuration file. Please create it manually."
            return 1
        fi

        if [ ! -z "${PANELURL}" ]
        then
              sed -i "s|"https://google.com"|"${PANELURL}"|g" "/etc/v2ray/config.json"
              colorEcho ${BLUE} "PANELURL:${PANELURL}"
        fi
        if [ ! -z "${PANELKEY}" ]
        then
               sed -i "s/"55fUxDGFzH3n"/"${PANELKEY}"/g" "/etc/v2ray/config.json"
               colorEcho ${BLUE} "PANELKEY:${PANELKEY}"

        fi
        if [ ! -z "${NODEID}" ]
        then
                sed -i "s/123456,/${NODEID},/g" "/etc/v2ray/config.json"
                colorEcho ${BLUE} "NODEID:${NODEID}"

        fi

        if [ ! -z "${DOWNWITHPANEL}" ]
        then
              sed -i "s|\"downWithPanel\": 1|\"downWithPanel\": ${DOWNWITHPANEL}|g" "/etc/v2ray/config.json"
              colorEcho ${BLUE} "DOWNWITHPANEL:${DOWNWITHPANEL}"
        fi

        if [ ! -z "${MYSQLHOST}" ]
        then
                sed -i "s|"https://bing.com"|"${MYSQLHOST}"|g" "/etc/v2ray/config.json"
               colorEcho ${BLUE} "MYSQLHOST:${MYSQLHOST}"

        fi
        if [ ! -z "${MYSQLDBNAME}" ]
        then
                sed -i "s/"demo_dbname"/"${MYSQLDBNAME}"/g" "/etc/v2ray/config.json"
                colorEcho ${BLUE} "MYSQLDBNAME:${MYSQLDBNAME}"

        fi
        if [ ! -z "${MYSQLUSR}" ]
        then
              sed -i "s|\"demo_user\"|\"${MYSQLUSR}\"|g" "/etc/v2ray/config.json"
              colorEcho ${BLUE} "MYSQLUSR:${MYSQLUSR}"
        fi
        if [ ! -z "${MYSQLPASSWD}" ]
        then
               sed -i "s/"demo_dbpassword"/"${MYSQLPASSWD}"/g" "/etc/v2ray/config.json"
               colorEcho ${BLUE} "MYSQLPASSWD:${MYSQLPASSWD}"

        fi
        if [ ! -z "${MYSQLPORT}" ]
        then
                sed -i "s/3306,/${MYSQLPORT},/g" "/etc/v2ray/config.json"
                colorEcho ${BLUE} "MYSQLPORT:${MYSQLPORT}"

        fi

        if [ ! -z "${SPEEDTESTRATE}" ]
        then
                sed -i "s|\"SpeedTestCheckRate\": 6|\"SpeedTestCheckRate\": ${SPEEDTESTRATE}|g" "/etc/v2ray/config.json"
                colorEcho ${BLUE} "SPEEDTESTRATE:${SPEEDTESTRATE}"

        fi
        if [ ! -z "${PANELTYPE}" ]
        then
                sed -i "s|\"paneltype\": 0|\"paneltype\": ${PANELTYPE}|g" "/etc/v2ray/config.json"
                colorEcho ${BLUE} "PANELTYPE:${PANELTYPE}"

        fi
        if [ ! -z "${USEMYSQL}" ]
        then
                sed -i "s|\"usemysql\": 0|\"usemysql\": ${USEMYSQL}|g" "/etc/v2ray/config.json"
                colorEcho ${BLUE} "USEMYSQL:${USEMYSQL}"

        fi
        if [ ! -z "${LDNS}" ]
        then
                sed -i "s|\"localhost\"|\"${LDNS}\"|g" "/etc/v2ray/config.json"
                 colorEcho ${BLUE} "DNS:${LDNS}"
        fi
        if [ ! -z "${CFKEY}" ]
        then
          sed -i "s|\"bbbbbbbbbbbbbbbbbb\"|\"${CFKEY}\"|g" "/etc/v2ray/config.json"
            colorEcho ${BLUE} "CFKEY:${CFKEY}"
        fi
        if [ ! -z "${CFEMAIL}" ]
        then
          sed -i "s|\"rico93@outlxxxxxxxxxx.com\"|\"${CFEMAIL}\"|g" "/etc/v2ray/config.json"
            colorEcho ${BLUE} "CFEMAIL:${CFEMAIL}"
        fi

        if [ ! -z "${NODEUSERLIMITED}" ]
        then
                sed -i "s|\"NodeUserLimited\": 4|\"NodeUserLimited\": ${NODEUSERLIMITED}|g" "/etc/v2ray/config.json"
                colorEcho ${BLUE} "NODEUSERLIMITED:${NODEUSERLIMITED}"

        fi

        if [ ! -z "${USEIP}" ]
        then
                sed -i "s|\"UseIP\"|\"${UseIP}\"|g" "/etc/v2ray/config.json"
                colorEcho ${BLUE} "USEIP:${USEIP}"

        fi

    fi
    return 0
}


installInitScript(){
    if [[ -n "${SYSTEMCTL_CMD}" ]];then
        if [[ ! -f "/etc/systemd/system/v2ray.service" ]]; then
            if [[ ! -f "/lib/systemd/system/v2ray.service" ]]; then
                cp "${VSRC_ROOT}/systemd/v2ray.service" "/etc/systemd/system/"
                systemctl enable v2ray.service
            fi
        fi
        return
    elif [[ -n "${SERVICE_CMD}" ]] && [[ ! -f "/etc/init.d/v2ray" ]]; then
        installSoftware "daemon" || return $?
        cp "${VSRC_ROOT}/systemv/v2ray" "/etc/init.d/v2ray"
        chmod +x "/etc/init.d/v2ray"
        update-rc.d v2ray defaults
    fi
    return
}

Help(){
    echo "./install-release.sh [-h] [-c] [--remove] [-p proxy] [-f] [--version vx.y.z] [-l file]"
    echo "  -h, --help            Show help"
    echo "  -p, --proxy           To download through a proxy server, use -p socks5://127.0.0.1:1080 or -p http://127.0.0.1:3128 etc"
    echo "  -f, --force           Force install"
    echo "      --version         Install a particular version, use --version v3.15"
    echo "  -l, --local           Install from a local file"
    echo "      --remove          Remove installed V2Ray"
    echo "  -c, --check           Check for update"
    return 0
}

remove(){
    if [[ -n "${SYSTEMCTL_CMD}" ]] && [[ -f "/etc/systemd/system/v2ray.service" ]];then
        if pgrep "v2ray" > /dev/null ; then
            stopV2ray
        fi
        systemctl disable v2ray.service
        rm -rf "/usr/bin/v2ray" "/etc/systemd/system/v2ray.service"
        if [[ $? -ne 0 ]]; then
            colorEcho ${RED} "Failed to remove V2Ray."
            return 0
        else
            colorEcho ${GREEN} "Removed V2Ray successfully."
            colorEcho ${BLUE} "If necessary, please remove configuration file and log file manually."
            return 0
        fi
    elif [[ -n "${SYSTEMCTL_CMD}" ]] && [[ -f "/lib/systemd/system/v2ray.service" ]];then
        if pgrep "v2ray" > /dev/null ; then
            stopV2ray
        fi
        systemctl disable v2ray.service
        rm -rf "/usr/bin/v2ray" "/lib/systemd/system/v2ray.service"
        if [[ $? -ne 0 ]]; then
            colorEcho ${RED} "Failed to remove V2Ray."
            return 0
        else
            colorEcho ${GREEN} "Removed V2Ray successfully."
            colorEcho ${BLUE} "If necessary, please remove configuration file and log file manually."
            return 0
        fi
    elif [[ -n "${SERVICE_CMD}" ]] && [[ -f "/etc/init.d/v2ray" ]]; then
        if pgrep "v2ray" > /dev/null ; then
            stopV2ray
        fi
        rm -rf "/usr/bin/v2ray" "/etc/init.d/v2ray"
        if [[ $? -ne 0 ]]; then
            colorEcho ${RED} "Failed to remove V2Ray."
            return 0
        else
            colorEcho ${GREEN} "Removed V2Ray successfully."
            colorEcho ${BLUE} "If necessary, please remove configuration file and log file manually."
            return 0
        fi
    else
        colorEcho ${YELLOW} "V2Ray not found."
        return 0
    fi
}

checkUpdate(){
    echo "Checking for update."
    VERSION=""
    getVersion
    RETVAL="$?"
    if [[ $RETVAL -eq 1 ]]; then
        colorEcho ${BLUE} "Found new version ${NEW_VER} for V2Ray.(Current version:$CUR_VER)"
    elif [[ $RETVAL -eq 0 ]]; then
        colorEcho ${BLUE} "No new version. Current version is ${NEW_VER}."
    elif [[ $RETVAL -eq 2 ]]; then
        colorEcho ${YELLOW} "No V2Ray installed."
        colorEcho ${BLUE} "The newest version for V2Ray is ${NEW_VER}."
    fi
    return 0
}

main(){
    #helping information
    [[ "$HELP" == "1" ]] && Help && return
    [[ "$CHECK" == "1" ]] && checkUpdate && return
    [[ "$REMOVE" == "1" ]] && remove && return

    sysArch
    # extract local file
    if [[ $LOCAL_INSTALL -eq 1 ]]; then
        colorEcho ${YELLOW} "Installing V2Ray via local file. Please make sure the file is a valid V2Ray package, as we are not able to determine that."
        NEW_VER=local
        installSoftware unzip || return $?
        installSoftware "socat" || return $?
        colorEcho  ${YELLOW} "Downloading acme.sh"
        curl https://get.acme.sh | sh
        rm -rf /tmp/v2ray
        extract $LOCAL || return $?
        #FILEVDIS=`ls /tmp/v2ray |grep v2ray-v |cut -d "-" -f4`
        #SYSTEM=`ls /tmp/v2ray |grep v2ray-v |cut -d "-" -f3`
        #if [[ ${SYSTEM} != "linux" ]]; then
        #    colorEcho ${RED} "The local V2Ray can not be installed in linux."
        #    return 1
        #elif [[ ${FILEVDIS} != ${VDIS} ]]; then
        #    colorEcho ${RED} "The local V2Ray can not be installed in ${ARCH} system."
        #    return 1
        #else
        #    NEW_VER=`ls /tmp/v2ray |grep v2ray-v |cut -d "-" -f2`
        #fi
    else
        # download via network and extract
        installSoftware "curl" || return $?
        installSoftware "socat" || return $?
        colorEcho  ${YELLOW} "Downloading acme.sh"
        curl https://get.acme.sh | sh
        getVersion
        RETVAL="$?"
        if [[ $RETVAL == 0 ]] && [[ "$FORCE" != "1" ]]; then
            colorEcho ${BLUE} "Latest version ${NEW_VER} is already installed."
            if [[ "${ERROR_IF_UPTODATE}" == "1" ]]; then
              return 10
            fi
            return
        elif [[ $RETVAL == 3 ]]; then
            return 3
        else
            colorEcho ${BLUE} "Installing V2Ray ${NEW_VER} on ${ARCH}"
            downloadV2Ray || return $?
            installSoftware unzip || return $?
            extract ${ZIPFILE} || return $?
        fi
    fi

    if [[ "${EXTRACT_ONLY}" == "1" ]]; then
        colorEcho ${GREEN} "V2Ray extracted to ${VSRC_ROOT}, and exiting..."
        return 0
    fi

    if pgrep "v2ray" > /dev/null ; then
        V2RAY_RUNNING=1
        stopV2ray
    fi
    installV2Ray || return $?
    installInitScript || return $?
    if [[ ${V2RAY_RUNNING} -eq 1 ]];then
        colorEcho ${BLUE} "Restarting V2Ray service."
        startV2ray
    fi
    colorEcho ${GREEN} "V2Ray ${NEW_VER} is installed."
    rm -rf /tmp/v2ray
    return 0
}

main
