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

func (m *MockStub) GetState(key string) ([]byte, error)    { return m.state[key], nil }
func (m *MockStub) PutState(key string, value []byte) error { m.state[key] = value; return nil }
func (m *MockStub) DelState(key string) error               { delete(m.state, key); return nil }
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
func (m *MockStub) GetCreator() ([]byte, error)                { return nil, nil }
func (m *MockStub) GetTransient() (map[string][]byte, error)   { return nil, nil }
func (m *MockStub) GetBinding() ([]byte, error)                { return nil, nil }
func (m *MockStub) GetDecorations() map[string][]byte          { return nil }
func (m *MockStub) GetSignedProposal() (*pb.SignedProposal, error) { return nil, nil }
func (m *MockStub) GetTxTimestamp() (*timestamp.Timestamp, error)  { return nil, nil }
func (m *MockStub) SetEvent(string, []byte) error                  { return nil }

type MockContext struct {
	stub *MockStub
}

func (c *MockContext) GetStub() shim.ChaincodeStubInterface  { return c.stub }
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

func seedAnonymizedKG(t *testing.T, stub *MockStub, kg AnonymizedKG) {
	t.Helper()
	b, err := json.Marshal(kg)
	if err != nil {
		t.Fatalf("failed to marshal anonymized KG: %v", err)
	}
	stub.PutState(kg.Id, b)
}

// --- StoreAnonymizedKG tests ---

func TestStoreAnonymizedKG_Success(t *testing.T) {
	cc := new(AnonymizedKGSmartContract)
	ctx := newMockContextWithCampaign(true)

	assertNoError(t, cc.StoreAnonymizedKG(ctx, "kg1", "c1", "recipient1", "rollup-envelope", "sig-abc"))

	raw, _ := ctx.stub.GetState("kg1")
	if raw == nil {
		t.Fatal("expected anonymized KG to be stored in ledger state")
	}
	var got AnonymizedKG
	json.Unmarshal(raw, &got)
	if got.Id != "kg1" || got.CampaignId != "c1" || got.RecipientId != "recipient1" {
		t.Fatalf("stored KG has unexpected values: %+v", got)
	}
	if got.Verified {
		t.Fatal("newly stored KG must not be verified")
	}
	if got.Shared {
		t.Fatal("newly stored KG must not be marked as shared")
	}
	if got.RecipientEnvelope != "" {
		t.Fatal("newly stored KG must have an empty recipient envelope")
	}
}

func TestStoreAnonymizedKG_DuplicateID(t *testing.T) {
	cc := new(AnonymizedKGSmartContract)
	ctx := newMockContextWithCampaign(true)

	assertNoError(t, cc.StoreAnonymizedKG(ctx, "kg1", "c1", "r1", "env", "sig"))
	assertError(t, cc.StoreAnonymizedKG(ctx, "kg1", "c1", "r1", "env", "sig"), "already exists")
}

func TestStoreAnonymizedKG_CampaignNotFound(t *testing.T) {
	cc := new(AnonymizedKGSmartContract)
	ctx := newMockContextWithCampaign(false)

	assertError(t, cc.StoreAnonymizedKG(ctx, "kg1", "missing-campaign", "r1", "env", "sig"), "")
}

// --- StoreProof tests ---

func TestStoreProof_Verified(t *testing.T) {
	cc := new(AnonymizedKGSmartContract)
	ctx := newMockContext()
	seedAnonymizedKG(t, ctx.stub, AnonymizedKG{Id: "kg1", CampaignId: "c1"})

	verified, err := cc.StoreProof(ctx, "kg1", "commit-abc", "commit-abc")
	assertNoError(t, err)
	if !verified {
		t.Fatal("expected StoreProof to return true when commits match")
	}

	var stored AnonymizedKG
	raw, _ := ctx.stub.GetState("kg1")
	json.Unmarshal(raw, &stored)
	if !stored.Verified {
		t.Fatal("expected KG.Verified to be true after matching commits")
	}
}

func TestStoreProof_NotVerified(t *testing.T) {
	cc := new(AnonymizedKGSmartContract)
	ctx := newMockContext()
	seedAnonymizedKG(t, ctx.stub, AnonymizedKG{Id: "kg1", CampaignId: "c1"})

	verified, err := cc.StoreProof(ctx, "kg1", "commit-user", "commit-rollup-different")
	assertNoError(t, err)
	if verified {
		t.Fatal("expected StoreProof to return false when commits do not match")
	}

	var stored AnonymizedKG
	raw, _ := ctx.stub.GetState("kg1")
	json.Unmarshal(raw, &stored)
	if stored.Verified {
		t.Fatal("expected KG.Verified to remain false when commits do not match")
	}
}

func TestStoreProof_NotFound(t *testing.T) {
	cc := new(AnonymizedKGSmartContract)
	ctx := newMockContext()

	_, err := cc.StoreProof(ctx, "nonexistent", "commit-a", "commit-b")
	assertError(t, err, "does not exist")
}

// --- ShareAnonymizedKGWithRecipient tests ---

func TestShareAnonymizedKGWithRecipient_Success(t *testing.T) {
	cc := new(AnonymizedKGSmartContract)
	ctx := newMockContextWithCampaign(true)
	seedAnonymizedKG(t, ctx.stub, AnonymizedKG{
		Id: "kg1", CampaignId: "c1", RecipientId: "r1",
		Verified: true, Shared: false,
	})

	assertNoError(t, cc.ShareAnonymizedKGWithRecipient(ctx, "kg1", "c1", "r1", "recipient-envelope"))

	var stored AnonymizedKG
	raw, _ := ctx.stub.GetState("kg1")
	json.Unmarshal(raw, &stored)
	if !stored.Shared {
		t.Fatal("expected KG.Shared to be true after sharing")
	}
	if stored.RecipientEnvelope != "recipient-envelope" {
		t.Fatalf("expected recipient envelope to be stored, got: %q", stored.RecipientEnvelope)
	}
}

func TestShareAnonymizedKGWithRecipient_NotVerified(t *testing.T) {
	cc := new(AnonymizedKGSmartContract)
	ctx := newMockContextWithCampaign(true)
	seedAnonymizedKG(t, ctx.stub, AnonymizedKG{
		Id: "kg1", CampaignId: "c1", RecipientId: "r1",
		Verified: false, Shared: false,
	})

	assertError(t, cc.ShareAnonymizedKGWithRecipient(ctx, "kg1", "c1", "r1", "envelope"), "not verified")
}

func TestShareAnonymizedKGWithRecipient_AlreadyShared(t *testing.T) {
	cc := new(AnonymizedKGSmartContract)
	ctx := newMockContextWithCampaign(true)
	seedAnonymizedKG(t, ctx.stub, AnonymizedKG{
		Id: "kg1", CampaignId: "c1", RecipientId: "r1",
		Verified: true, Shared: true,
	})

	assertError(t, cc.ShareAnonymizedKGWithRecipient(ctx, "kg1", "c1", "r1", "envelope"), "already shared")
}

func TestShareAnonymizedKGWithRecipient_WrongRecipient(t *testing.T) {
	cc := new(AnonymizedKGSmartContract)
	ctx := newMockContextWithCampaign(true)
	seedAnonymizedKG(t, ctx.stub, AnonymizedKG{
		Id: "kg1", CampaignId: "c1", RecipientId: "r1",
		Verified: true, Shared: false,
	})

	assertError(t, cc.ShareAnonymizedKGWithRecipient(ctx, "kg1", "c1", "wrong-recipient", "envelope"), "wrong recipient")
}

func TestShareAnonymizedKGWithRecipient_CampaignNotFound(t *testing.T) {
	cc := new(AnonymizedKGSmartContract)
	ctx := newMockContextWithCampaign(false)
	seedAnonymizedKG(t, ctx.stub, AnonymizedKG{
		Id: "kg1", CampaignId: "c1", RecipientId: "r1",
		Verified: true, Shared: false,
	})

	assertError(t, cc.ShareAnonymizedKGWithRecipient(ctx, "kg1", "c1", "r1", "envelope"), "does not exist")
}

// --- DeleteAnonymizedKG tests ---

func TestDeleteAnonymizedKG_Success(t *testing.T) {
	cc := new(AnonymizedKGSmartContract)
	ctx := newMockContext()
	seedAnonymizedKG(t, ctx.stub, AnonymizedKG{Id: "kg1"})

	assertNoError(t, cc.DeleteAnonymizedKG(ctx, "kg1"))

	raw, _ := ctx.stub.GetState("kg1")
	if raw != nil {
		t.Fatal("expected KG to be deleted from ledger state")
	}
}

func TestDeleteAnonymizedKG_NotFound(t *testing.T) {
	cc := new(AnonymizedKGSmartContract)
	ctx := newMockContext()

	assertError(t, cc.DeleteAnonymizedKG(ctx, "nonexistent"), "does not exist")
}
