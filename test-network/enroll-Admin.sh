#!/bin/sh
  
fabric-ca-client register --id.name GFMAdmin --id.secret GFMAdminpw --id.type client --id.affiliation org1 --id.attrs 'nodeName=GFMAdmin:' --tls.certfiles "${PWD}/organizations/fabric-ca/org1/tls-cert.pem"

fabric-ca-client enroll -u https://GFMAdmin:GFMAdminpw@localhost:7054 --caname ca-org1 --enrollment.attrs "nodeName" -M "${PWD}/organizations/peerOrganizations/org1.example.com/users/GFMAdmin@org1.example.com/msp" --tls.certfiles "${PWD}/organizations/fabric-ca/org1/tls-cert.pem"

cp "${PWD}/organizations/peerOrganizations/org1.example.com/msp/config.yaml" "${PWD}/organizations/peerOrganizations/org1.example.com/users/GFMAdmin@org1.example.com/msp/config.yaml"
