#!/bin/bash

function one_line_pem {
    echo "`awk 'NF {sub(/\\n/, ""); printf "%s\\\\\\\n",$0;}' $1`"
}

function yaml_ccp {
    local PP=$(one_line_pem $4)
    local CP=$(one_line_pem $5)
    sed -e "s/\${ORG}/$1/" \
        -e "s/\${PEER0_PORT}/$2/" \
        -e "s/\${CAPORT}/$3/" \
        -e "s#\${PEERPEM}#$PP#" \
        -e "s#\${CAPEM}#$CP#" \
        $SCRIPTS_DIR/utils/ccp-template.yaml | sed -e $'s/\\\\n/\\\n          /g'
}

ORG="rec0"
PEER0_PORT=1050
CAPORT=7054
PEERPEM=organizations/peerOrganizations/rec0.tracenet.com/tlsca/tlsca.rec0.tracenet.com-cert.pem
CAPEM=organizations/peerOrganizations/rec0.tracenet.com/ca/ca.rec0.tracenet.com-cert.pem

echo "$(yaml_ccp $ORG $PEER0_PORT $CAPORT $PEERPEM $CAPEM)" > organizations/peerOrganizations/rec0.tracenet.com/connection-rec0.yaml

ORG="obs0"
PEER0_PORT=2050
CAPORT=8054
PEERPEM=organizations/peerOrganizations/obs0.tracenet.com/tlsca/tlsca.obs0.tracenet.com-cert.pem
CAPEM=organizations/peerOrganizations/obs0.tracenet.com/ca/ca.obs0.tracenet.com-cert.pem

ORG="obs0"
PEER0_PORT=2050
CAPORT=8054
PEERPEM=organizations/peerOrganizations/obs0.tracenet.com/tlsca/tlsca.obs0.tracenet.com-cert.pem
CAPEM=organizations/peerOrganizations/obs0.tracenet.com/ca/ca.obs0.tracenet.com-cert.pem

ORG="prov0"
PEER0_PORT=3050
CAPORT=8054
PEERPEM=organizations/peerOrganizations/prov0.tracenet.com/tlsca/tlsca.prov0.tracenet.com-cert.pem
CAPEM=organizations/peerOrganizations/prov0.tracenet.com/ca/ca.prov0.tracenet.com-cert.pem

echo "$(yaml_ccp $ORG $PEER0_PORT $CAPORT $PEERPEM $CAPEM)" > organizations/peerOrganizations/obs0.tracenet.com/connection-obs0.yaml
