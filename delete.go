// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// File contains Add functionality
//
// https://tools.ietf.org/html/rfc4511
//
// DelRequest ::= [APPLICATION 10] LDAPDN
//
//

package ldap

import (
	"errors"
	"gopkg.in/asn1-ber.v1"
	"log"
)

type DeleteRequest struct {
	dn string
}

func (d DeleteRequest) encode() *ber.Packet {
	//request := ber.Encode(ber.ClassApplication, ber.TypePrimitive, ApplicationDelRequest, nil, "Del Request")
	//request.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, d.dn, "DN"))
	request := ber.NewString(ber.ClassApplication, ber.TypePrimitive, ApplicationDelRequest, d.dn, "Del Request")
	return request
}

func NewDeleteRequest(
	dn string,
) *DeleteRequest {
	return &DeleteRequest{
		dn: dn,
	}
}

func (l *Conn) Delete(deleteRequest *DeleteRequest) error {
	messageID := l.nextMessageID()
	packet := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "LDAP Request")
	packet.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, messageID, "MessageID"))
	packet.AppendChild(deleteRequest.encode())

	l.Debug.PrintPacket(packet)

	channel, err := l.sendMessage(packet)
	if err != nil {
		return err
	}
	if channel == nil {
		return NewError(ErrorNetwork, errors.New("ldap: could not send message"))
	}
	defer l.finishMessage(messageID)

	l.Debug.Printf("%d: waiting for response", messageID)
	packet = <-channel
	l.Debug.Printf("%d: got response %p", messageID, packet)
	if packet == nil {
		return NewError(ErrorNetwork, errors.New("ldap: could not retrieve message"))
	}

	if l.Debug {
		if err := addLDAPDescriptions(packet); err != nil {
			return err
		}
		ber.PrintPacket(packet)
	}

	if packet.Children[1].Tag == ApplicationDelResponse {
		resultCode, resultDescription := getLDAPResultCode(packet)
		if resultCode != 0 {
			return NewError(resultCode, errors.New(resultDescription))
		}
	} else {
		log.Printf("Unexpected Response: %d", packet.Children[1].Tag)
	}

	l.Debug.Printf("%d: returning", messageID)
	return nil
}
