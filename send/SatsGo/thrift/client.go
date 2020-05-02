package thrift

import (
    "fmt"
    "sync"
)

type TStandardClient struct {
	seqId        int32
	iprot, oprot TProtocol
    mutex  *sync.Mutex
}

// TStandardClient implements TClient, and uses the standard message format for Thrift.
// It is not safe for concurrent use.
func NewTStandardClient(inputProtocol, outputProtocol TProtocol) *TStandardClient {
	return &TStandardClient{
		iprot: inputProtocol,
		oprot: outputProtocol,
        mutex: new(sync.Mutex),
	}
}

func (p *TStandardClient) Send(oprot TProtocol, seqId int32, method string, args TStruct) error {
	if err := oprot.WriteMessageBegin(method, CALL, seqId); err != nil {
		return err
	}
	if err := args.Write(oprot); err != nil {
		return err
	}
	if err := oprot.WriteMessageEnd(); err != nil {
		return err
	}
	return oprot.Flush()
}

func (p *TStandardClient) Recv(iprot TProtocol, seqId int32, method string, result TStruct) error {
	rMethod, rTypeId, rSeqId, err := iprot.ReadMessageBegin()
	if err != nil {
		return err
	}

	if method != rMethod {
		return NewTApplicationException(WRONG_METHOD_NAME, fmt.Sprintf("%s: wrong method name", method))
	} else if seqId != rSeqId {
        return nil
		//return NewTApplicationException(BAD_SEQUENCE_ID, fmt.Sprintf("%s: out of order sequence response", method))
	} else if rTypeId == EXCEPTION {
		var exception tApplicationException
		if err := exception.Read(iprot); err != nil {
			return err
		}

		if err := iprot.ReadMessageEnd(); err != nil {
			return err
		}
		return &exception
	} else if rTypeId != REPLY {
		return NewTApplicationException(INVALID_MESSAGE_TYPE_EXCEPTION, fmt.Sprintf("%s: invalid message type", method))
	}

	if err := result.Read(iprot); err != nil {
		return err
	}
	
	return iprot.ReadMessageEnd()
}

func (p *TStandardClient) call(method string, args, result TStruct) error {
	
	p.mutex.Lock()
	p.seqId++
	seqId := p.seqId
	if err := p.Send(p.oprot, seqId, method, args); err != nil {
		p.mutex.Unlock()
		return err
	}

	// method is oneway
	if result == nil {
		p.mutex.Unlock()
		return nil
	}

	r := p.Recv(p.iprot, seqId, method, result)
	p.mutex.Unlock()
	return r
}
