#!/bin/sh
  
fabric-ca-client register --id.name SMBroker --id.secret SMBrokerpw --id.type client --id.affiliation org1 --id.attrs 'nodeName=SMBroker:' --tls.certfiles "${PWD}/organizations/fabric-ca/org1/tls-cert.pem"

fabric-ca-client enroll -u https://SMBroker:SMBrokerpw@localhost:7054 --caname ca-org1 --enrollment.attrs "nodeName" -M "${PWD}/organizations/peerOrganizations/org1.example.com/users/SMBroker@org1.example.com/msp" --tls.certfiles "${PWD}/organizations/fabric-ca/org1/tls-cert.pem"

cp "${PWD}/organizations/peerOrganizations/org1.example.com/msp/config.yaml" "${PWD}/organizations/peerOrganizations/org1.example.com/users/SMBroker@org1.example.com/msp/config.yaml"
