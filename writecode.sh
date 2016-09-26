#!/bin/bash

tables=(
    hosts
    services
    servicesets
    hosttemplates
    servicetemplates
    hostgroups
    servicegroups
    contacts
    contactgroups
    timeperiods
    commands
    servicedeps
    hostdeps
    serviceesc
    hostesc
    serviceextinfo
    hostextinfo
)

hosts="
        name alias ipaddress template hostgroup contact contactgroups
        activechecks servicesets disable displayname parents command
        initialstate maxcheckattempts checkinterval retryinterval passivechecks
        checkperiod obsessoverhost checkfreshness freshnessthresh eventhandler
        eventhandlerenabled lowflapthresh highflapthresh flapdetectionenabled
        flapdetectionoptions processperfdata retainstatusinfo
        retainnonstatusinfo notifinterval firstnotifdelay notifperiod notifopts
        notifications_enabled stalkingoptions notes notes_url icon_image
        icon_image_alt vrml_image statusmap_image coords2d coords3d action_url
        customvars
"
hosts_encode="command alias"
hosts_required="name alias ipaddress template"

services="
        name template command svcdesc svcgroup contacts contactgroups
        freshnessthresh activechecks customvars disable displayname isvolatile
        initialstate maxcheckattempts checkinterval retryinterval passivechecks
        checkperiod obsessoverservice manfreshnessthresh checkfreshness
        eventhandler eventhandlerenabled lowflapthresh highflapthresh
        flapdetectionenabled flapdetectionoptions processperfdata
        retainstatusinfo retainnonstatusinfo notifinterval firstnotifdelay
        notifperiod notifopts notifications_enabled stalkingoptions notes
        notes_url action_url icon_image icon_image_alt vrml_image
        statusmap_image coords2d coords3d
"
services_encode="name command svcdesc"
services_required="name template command svcdesc"

servicesets="
        name template command svcdesc svcgroup contacts contactgroups
        freshnessthresh activechecks customvars disable displayname isvolatile
        initialstate maxcheckattempts checkinterval retryinterval passivechecks
        checkperiod obsessoverservice manfreshnessthresh checkfreshness
        eventhandler eventhandlerenabled lowflapthresh highflapthresh
        flapdetectionenabled flapdetectionoptions processperfdata
        retainstatusinfo retainnonstatusinfo notifinterval firstnotifdelay
        notifperiod notifopts notifications_enabled stalkingoptions notes
        notes_url action_url icon_image icon_image_alt vrml_image
        statusmap_image coords2d coords3d
"
servicesets_encode="name command svcdesc"
servicesets_required="name template command svcdesc"

hosttemplates="
        name use contacts contactgroups normchecki checkinterval retryinterval
        notifperiod notifopts disable checkperiod maxcheckattempts checkcommand
        notifinterval passivechecks obsessoverhost checkfreshness
        freshnessthresh eventhandler eventhandlerenabled lowflapthresh
        highflapthresh flapdetectionenabled flapdetectionoptions
        processperfdata retainstatusinfo retainnonstatusinfo firstnotifdelay
        notifications_enabled stalkingoptions notes notes_url icon_image
        icon_image_alt vrml_image statusmap_image coords2d coords3d action_url
        customvars
"
hosttemplates_encode="checkcommand action_url"
hosttemplates_required="name checkinterval retryinterval notifperiod checkperiod maxcheckattempts notifinterval"

servicetemplates="
        name use contacts contactgroups notifopts checkinterval normchecki
        retryinterval notifinterval notifperiod disable checkperiod
        maxcheckattempts freshnessthresh activechecks customvars isvolatile
        initialstate passivechecks obsessoverservice manfreshnessthresh
        checkfreshness eventhandler eventhandlerenabled lowflapthresh
        highflapthresh flapdetectionenabled flapdetectionoptions
        processperfdata retainstatusinfo retainnonstatusinfo firstnotifdelay
        notifications_enabled stalkingoptions notes notes_url action_url
        icon_image icon_image_alt vrml_image statusmap_image coords2d coords3d
"
servicetemplates_encode="action_url"
servicetemplates_required="name checkinterval retryinterval notifinterval notifperiod checkperiod maxcheckattempts"

hostgroups="
        name alias disable members hostgroupmembers notes notes_url action_url
"
hostgroups_encode=""
hostgroups_required="name alias"

servicegroups="
        name alias disable members servicegroupmembers notes notes_url
        action_url
"
servicegroups_encode=""
servicegroups_required="name alias"

contacts="
        name use alias emailaddr svcnotifperiod svcnotifopts svcnotifcmds
        hstnotifperiod hstnotifopts hstnotifcmds cansubmitcmds disable
        svcnotifenabled hstnotifenabled pager address1 address2 address3
        address4 address5 address6 retainstatusinfo retainnonstatusinfo
        contactgroups
"
contacts_encode=""
contacts_required="name alias svcnotifperiod svcnotifopts svcnotifcmds hstnotifperiod hstnotifopts hstnotifcmds"

contactgroups="
        name alias members disable
"
contactgroups_encode=""
contactgroups_required="name alias members"

timeperiods="
        name alias definition exclude disable exception
"
timeperiods_encode=""
timeperiods_required="name alias"

commands="
        name command disable
"
commands_encode="name command"
commands_required="name command"

servicedeps="
        dephostname dephostgroupname depsvcdesc hostname hostgroupname svcdesc
        inheritsparent execfailcriteria notiffailcriteria period disable
"
servicedeps_encode=""
servicedeps_required=""

hostdeps="
        dephostname dephostgroupname hostname hostgroupname inheritsparent
        execfailcriteria notiffailcriteria period disable
"
hostdeps_encode=""
hostdeps_required=""

serviceesc="
        hostname hostgroupname svcdesc contacts contactgroups firstnotif
        lastnotif notifinterval period escopts disable
"
serviceesc_encode=""
serviceesc_required=""

hostesc="
        hostname hostgroupname contacts contactgroups firstnotif lastnotif
        notifinterval period escopts disable
"
hostesc_encode=""
hostesc_required=""

serviceextinfo="
        hostname svcdesc notes notes_url action_url icon_image icon_image_alt
        disable
"
serviceextinfo_encode=""
serviceextinfo_required=""

hostextinfo="
        hostname notes notes_url action_url icon_image icon_image_alt
        vrml_image statusmap_image coords2d coords3d disable
"
hostextinfo_encode=""
hostextinfo_required=""

mkdir -p newfiles

TMPFILE=/tmp/writecode.sh.tmp
NEWGO=/tmp/writecode.go.tmp

for table in ${tables[*]}; do
    a=`cut -b1 <<<$table | tr [a-z] [A-Z]`
    b=`cut -b2- <<<$table`
    rHosts="$a$b"
    rhosts="$table"
    rhost="${table%s}"
    sed "s/%Hosts%/$rHosts/g;s/%hosts%/$rhosts/g;s/%host%/$rhost/g" \
        template.tmpl >$NEWGO
    :>$TMPFILE
    for i in $(eval echo $`echo $table`); do
        echo -e "\t\t$i\t\t\tstring" >>$TMPFILE
    done
    sed -i "/%host_contents%/ r $TMPFILE" $NEWGO
    sed -i "/%host_contents%/d" $NEWGO
    :>$TMPFILE
    for i in $(eval echo $`echo $table`); do
        encode=0
        for j in $(eval echo $`echo ${table}_encode`); do
            [[ $i == $j ]] && encode=1
        done
        if [[ encode -eq 0 ]]; then
            echo -e "case \"$i\":" >>$TMPFILE
            echo -e "\t $rhost.$i = v.(string)" >>$TMPFILE
        else
            echo -e "case \"$i\":" >>$TMPFILE
            echo -e "\t $rhost.$i, _ = UrlDecode(v.(string))" >>$TMPFILE
        fi
    done
    sed -i "/%host_case_content%/ r $TMPFILE" $NEWGO
    sed -i "/%host_case_content%/d" $NEWGO
    #
    c=""
    items=""
    for i in $(eval echo $`echo ${table}_required`); do
        items+="$c\"$i\""
        c=","
    done
    sed -i "s/%reqfields%/\treturn []string{$items}/" $NEWGO
    #
    rm -f $TMPFILE
    gofmt -w $NEWGO
    [[ $? -ne 0 ]] && exit 1
    mv $NEWGO newfiles/$table.go
done
