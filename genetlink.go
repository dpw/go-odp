package main

import (
        "syscall"
	"fmt"
	"unsafe"
)

func (nlmsg *NlMsgBuilder) AddGenlMsghdr(cmd uint8) (res *GenlMsghdr) {
	nlmsg.Align(syscall.NLMSG_ALIGNTO)
	pos := nlmsg.Grow(unsafe.Sizeof(*res))
	res = genlMsghdrAt(nlmsg.buf, pos)
	res.Cmd = cmd
	return
}

func (nlmsg *NlMsgButcher) TakeGenlMsghdr(expectCmd uint8) (*GenlMsghdr, error) {
	nlmsg.Align(syscall.NLMSG_ALIGNTO)
	gh := genlMsghdrAt(nlmsg.data, nlmsg.pos)
	if err := nlmsg.Advance(unsafe.Sizeof(*gh)); err != nil {
		return nil, err
	}

	if gh.Cmd != expectCmd {
		return nil, fmt.Errorf("generic netlink response has wrong cmd (got %d, expected %d)", gh.Cmd, expectCmd)
	}

	return gh, nil
}

type GenlFamilyId uint16

func (s *NetlinkSocket) LookupGenlFamily(name string) (GenlFamilyId, error) {
	req := NewNlMsgBuilder(syscall.NLM_F_REQUEST, GENL_ID_CTRL)

	req.AddGenlMsghdr(CTRL_CMD_GETFAMILY)
	req.AddStringRtAttr(CTRL_ATTR_FAMILY_NAME, name)
	b, seq := req.Finish()

	if err := s.send(b); err != nil {
		return 0, err
        }

	rb, err := s.recv(0)
        if err != nil {
		return 0, err
        }

	if err = s.checkResponse(rb, seq); err != nil {
		return 0, err
	}

	resp := NewNlMsgButcher(rb)
	if _, err := resp.TakeNlMsghdr(GENL_ID_CTRL); err != nil {
		return 0, err
	}

	if _, err := resp.TakeGenlMsghdr(CTRL_CMD_NEWFAMILY); err != nil {
		return 0, err
	}

	attrs, err := resp.TakeAttrs()
	if err != nil {
		return 0, err
	}

	id, err := attrs.GetUint16(CTRL_ATTR_FAMILY_ID)
	if err != nil {
		return 0, err
	}

	return GenlFamilyId(id), nil
}
