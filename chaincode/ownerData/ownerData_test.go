package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

// --- Mock infrastructure ---

type MockStub struct {
	state               map[string][]byte
	invokeChaincodeFunc func(string, [][]byte, string) pb.Response
}

func newMockStub() *MockStub {
	return &MockStub{state: make(map[string][]byte)}
}

func (m *MockStub) GetState(key string) ([]byte, error)   { return m.state[key], nil }
func (m *MockStub) PutState(key string, value []byte) error { m.state[key] = value; return nil }
func (m *MockStub) DelState(key string) error              { delete(m.state, key); return nil }
func (m *MockStub) InvokeChaincode(name string, args [][]byte, channel string) pb.Response {
	if m.invokeChaincodeFunc != nil {
		return m.invokeChaincodeFunc(name, args, channel)
	}
	return pb.Response{}
}
func (m *MockStub) GetArgs() [][]byte                                       { return nil }
func (m *MockStub) GetStringArgs() []string                                 { return nil }
func (m *MockStub) GetFunctionAndParameters() (string, []string)            { return "", nil }
func (m *MockStub) GetArgsSlice() ([]byte, error)                           { return nil, nil }
func (m *MockStub) GetTxID() string                                         { return "mock-tx-id" }
func (m *MockStub) GetChannelID() string                                    { return "mychannel" }
func (m *MockStub) SetStateValidationParameter(string, []byte) error        { return nil }
func (m *MockStub) GetStateValidationParameter(string) ([]byte, error)      { return nil, nil }
func (m *MockStub) GetStateByRange(string, string) (shim.StateQueryIteratorInterface, error) {
	return nil, nil
}
func (m *MockStub) GetStateByRangeWithPagination(string, string, int32, string) (shim.StateQueryIteratorInterface, *pb.QueryResponseMetadata, error) {
	return nil, nil, nil
}
func (m *MockStub) GetStateByPartialCompositeKey(string, []string) (shim.StateQueryIteratorInterface, error) {
	return nil, nil
}
func (m *MockStub) GetStateByPartialCompositeKeyWithPagination(string, []string, int32, string) (shim.StateQueryIteratorInterface, *pb.QueryResponseMetadata, error) {
	return nil, nil, nil
}
func (m *MockStub) CreateCompositeKey(string, []string) (string, error)    { return "", nil }
func (m *MockStub) SplitCompositeKey(string) (string, []string, error)     { return "", nil, nil }
func (m *MockStub) GetQueryResult(string) (shim.StateQueryIteratorInterface, error) {
	return nil, nil
}
func (m *MockStub) GetQueryResultWithPagination(string, int32, string) (shim.StateQueryIteratorInterface, *pb.QueryResponseMetadata, error) {
	return nil, nil, nil
}
func (m *MockStub) GetHistoryForKey(string) (shim.HistoryQueryIteratorInterface, error) {
	return nil, nil
}
func (m *MockStub) GetPrivateData(string, string) ([]byte, error)                    { return nil, nil }
func (m *MockStub) GetPrivateDataHash(string, string) ([]byte, error)                { return nil, nil }
func (m *MockStub) PutPrivateData(string, string, []byte) error                      { return nil }
func (m *MockStub) DelPrivateData(string, string) error                              { return nil }
func (m *MockStub) PurgePrivateData(string, string) error                            { return nil }
func (m *MockStub) SetPrivateDataValidationParameter(string, string, []byte) error   { return nil }
func (m *MockStub) GetPrivateDataValidationParameter(string, string) ([]byte, error) { return nil, nil }
func (m *MockStub) GetPrivateDataByRange(string, string, string) (shim.StateQueryIteratorInterface, error) {
	return nil, nil
}
func (m *MockStub) GetPrivateDataByPartialCompositeKey(string, string, []string) (shim.StateQueryIteratorInterface, error) {
	return nil, nil
}
func (m *MockStub) GetPrivateDataQueryResult(string, string) (shim.StateQueryIteratorInterface, error) {
	return nil, nil
}
func (m *MockStub) GetCreator() ([]byte, error)               { return nil, nil }
func (m *MockStub) GetTransient() (map[string][]byte, error)  { return nil, nil }
func (m *MockStub) GetBinding() ([]byte, error)               { return nil, nil }
func (m *MockStub) GetDecorations() map[string][]byte         { return nil }
func (m *MockStub) GetSignedProposal() (*pb.SignedProposal, error) { return nil, nil }
func (m *MockStub) GetTxTimestamp() (*timestamp.Timestamp, error)  { return nil, nil }
func (m *MockStub) SetEvent(string, []byte) error                  { return nil }

type MockContext struct {
	stub *MockStub
}

func (c *MockContext) GetStub() shim.ChaincodeStubInterface { return c.stub }
func (c *MockContext) GetClientIdentity() cid.ClientIdentity { return nil }

func newMockContext() *MockContext {
	return &MockContext{stub: newMockStub()}
}

func newMockContextWithCampaign(exists bool) *MockContext {
	ctx := newMockContext()
	ctx.stub.invokeChaincodeFunc = func(name string, args [][]byte, channel string) pb.Response {
		if exists {
			return shim.Success([]byte("true"))
		}
		return shim.Success([]byte("false"))
	}
	return ctx
}

// --- Helpers ---

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertError(t *testing.T, err error, contains string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error containing %q but got nil", contains)
	}
	if contains != "" && !strings.Contains(err.Error(), contains) {
		t.Fatalf("expected error to contain %q, got: %v", contains, err)
	}
}

func seedOwnerData(t *testing.T, stub *MockStub, d OwnerData) {
	t.Helper()
	b, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("failed to marshal owner data: %v", err)
	}
	stub.PutState(d.Id, b)
}

// --- Tests ---

func TestShareData_Success(t *testing.T) {
	cc := new(OwnerDataSmartContract)
	ctx := newMockContextWithCampaign(true)

	assertNoError(t, cc.ShareData(ctx, "d1", "c1", "encrypted-envelope", "k=5"))

	raw, _ := ctx.stub.GetState("d1")
	if raw == nil {
		t.Fatal("expected owner data to be stored in ledger state")
	}
	var got OwnerData
	json.Unmarshal(raw, &got)
	if got.Id != "d1" || got.CampaignId != "c1" || got.PrivacyPreference != "k=5" {
		t.Fatalf("stored owner data has unexpected values: %+v", got)
	}
}

func TestShareData_DuplicateID(t *testing.T) {
	cc := new(OwnerDataSmartContract)
	ctx := newMockContextWithCampaign(true)

	assertNoError(t, cc.ShareData(ctx, "d1", "c1", "envelope", "k=5"))
	assertError(t, cc.ShareData(ctx, "d1", "c1", "envelope2", "k=10"), "already exists")
}

func TestShareData_CampaignNotFound(t *testing.T) {
	cc := new(OwnerDataSmartContract)
	ctx := newMockContextWithCampaign(false)

	assertError(t, cc.ShareData(ctx, "d1", "missing-campaign", "envelope", "k=5"), "")
}

func TestDeleteSharedData_Success(t *testing.T) {
	cc := new(OwnerDataSmartContract)
	ctx := newMockContext()
	seedOwnerData(t, ctx.stub, OwnerData{Id: "d1", CampaignId: "c1"})

	assertNoError(t, cc.DeleteSharedData(ctx, "d1"))

	raw, _ := ctx.stub.GetState("d1")
	if raw != nil {
		t.Fatal("expected owner data to be deleted from ledger state")
	}
}

func TestDeleteSharedData_NotFound(t *testing.T) {
	cc := new(OwnerDataSmartContract)
	ctx := newMockContext()

	assertError(t, cc.DeleteSharedData(ctx, "nonexistent"), "does not exist")
}
