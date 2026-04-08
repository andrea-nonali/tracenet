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
	state map[string][]byte
}

func newMockStub() *MockStub {
	return &MockStub{state: make(map[string][]byte)}
}

func (m *MockStub) GetState(key string) ([]byte, error)                    { return m.state[key], nil }
func (m *MockStub) PutState(key string, value []byte) error                { m.state[key] = value; return nil }
func (m *MockStub) DelState(key string) error                              { delete(m.state, key); return nil }
func (m *MockStub) InvokeChaincode(string, [][]byte, string) pb.Response   { return pb.Response{} }
func (m *MockStub) GetArgs() [][]byte                                      { return nil }
func (m *MockStub) GetStringArgs() []string                                { return nil }
func (m *MockStub) GetFunctionAndParameters() (string, []string)           { return "", nil }
func (m *MockStub) GetArgsSlice() ([]byte, error)                          { return nil, nil }
func (m *MockStub) GetTxID() string                                        { return "mock-tx-id" }
func (m *MockStub) GetChannelID() string                                   { return "mychannel" }
func (m *MockStub) SetStateValidationParameter(string, []byte) error       { return nil }
func (m *MockStub) GetStateValidationParameter(string) ([]byte, error)     { return nil, nil }
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

func seedCampaign(t *testing.T, stub *MockStub, c Campaign) {
	t.Helper()
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("failed to marshal campaign: %v", err)
	}
	stub.PutState(c.Id, b)
}

// --- Tests ---

func TestCreateCampaign_Success(t *testing.T) {
	cc := new(CampaignSmartContract)
	ctx := newMockContext()

	assertNoError(t, cc.CreateCampaign(ctx, "c1", "AML Data Collection", "1700000000", "1800000000"))

	raw, _ := ctx.stub.GetState("c1")
	if raw == nil {
		t.Fatal("expected campaign to be stored in ledger state")
	}
	var got Campaign
	json.Unmarshal(raw, &got)
	if got.Id != "c1" || got.Name != "AML Data Collection" {
		t.Fatalf("stored campaign has unexpected values: %+v", got)
	}
}

func TestCreateCampaign_DuplicateID(t *testing.T) {
	cc := new(CampaignSmartContract)
	ctx := newMockContext()

	assertNoError(t, cc.CreateCampaign(ctx, "c1", "First", "1000", "2000"))
	assertError(t, cc.CreateCampaign(ctx, "c1", "Duplicate", "1000", "2000"), "already exists")
}

func TestCampaignExists_True(t *testing.T) {
	cc := new(CampaignSmartContract)
	ctx := newMockContext()
	seedCampaign(t, ctx.stub, Campaign{Id: "c1", Name: "KYC Campaign"})

	exists, err := cc.CampaignExists(ctx, "c1")
	assertNoError(t, err)
	if !exists {
		t.Fatal("expected CampaignExists to return true")
	}
}

func TestCampaignExists_False(t *testing.T) {
	cc := new(CampaignSmartContract)
	ctx := newMockContext()

	exists, err := cc.CampaignExists(ctx, "nonexistent")
	assertNoError(t, err)
	if exists {
		t.Fatal("expected CampaignExists to return false for unknown ID")
	}
}

func TestQueryCampaign_ReturnsStoredData(t *testing.T) {
	cc := new(CampaignSmartContract)
	ctx := newMockContext()
	seedCampaign(t, ctx.stub, Campaign{Id: "c1", Name: "Credit Risk Q3", StartTime: "100", EndTime: "200"})

	got := cc.QueryCampaign(ctx, "c1")
	if got.Id != "c1" || got.Name != "Credit Risk Q3" || got.StartTime != "100" || got.EndTime != "200" {
		t.Fatalf("QueryCampaign returned unexpected data: %+v", got)
	}
}

func TestDeleteCampaign_Success(t *testing.T) {
	cc := new(CampaignSmartContract)
	ctx := newMockContext()
	seedCampaign(t, ctx.stub, Campaign{Id: "c1"})

	assertNoError(t, cc.DeleteCampaign(ctx, "c1"))

	exists, _ := cc.CampaignExists(ctx, "c1")
	if exists {
		t.Fatal("expected campaign to be deleted from ledger state")
	}
}

func TestDeleteCampaign_NotFound(t *testing.T) {
	cc := new(CampaignSmartContract)
	ctx := newMockContext()

	assertError(t, cc.DeleteCampaign(ctx, "nonexistent"), "does not exist")
}
