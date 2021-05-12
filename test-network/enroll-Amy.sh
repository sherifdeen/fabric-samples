#!/bin/sh
  
fabric-ca-client register --id.name Amy --id.secret Amypw --id.type client --id.affiliation org1 --id.attrs 'nodeName=Amy:' --tls.certfiles "${PWD}/organizations/fabric-ca/org1/tls-cert.pem"

fabric-ca-client enroll -u https://Amy:Amypw@localhost:7054 --caname ca-org1 --enrollment.attrs "nodeName" -M "${PWD}/organizations/peerOrganizations/org1.example.com/users/Amy@org1.example.com/msp" --tls.certfiles "${PWD}/organizations/fabric-ca/org1/tls-cert.pem"

cp "${PWD}/organizations/peerOrganizations/org1.example.com/msp/config.yaml" "${PWD}/organizations/peerOrganizations/org1.example.com/users/Amy@org1.example.com/msp/config.yaml"
