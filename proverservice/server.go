package main

import (
	"context"
	"fmt"
	"io"
	"math/big"
	"strconv"
	"time"

	"github.com/hermeznetwork/hermez-core/proverservice/pb"
)

type zkProverServiceServer struct {
	pb.ZKProverServiceServer
	id         int
	idsToState map[string]int
}

const (
	serverProtoVersion = "1"
	serverVersion      = "1"
)

var mockProof = &pb.Proof{
	ProofA: []string{"0", "0"},
	ProofB: []*pb.ProofB{{Proofs: []string{"0", "0"}}, {Proofs: []string{"0", "0"}}},
	ProofC: []string{"0", "0"},
}

func NewZkProverServiceServer() *zkProverServiceServer {
	idsToState := make(map[string]int)
	return &zkProverServiceServer{
		id:         0,
		idsToState: idsToState,
	}
}

func (zkp *zkProverServiceServer) GenProof(ctx context.Context, request *pb.GenProofRequest) (*pb.GenProofResponse, error) {
	zkp.id++
	idStr := strconv.Itoa(zkp.id)
	zkp.idsToState[idStr] = 0
	return &pb.GenProofResponse{
		Id:     idStr,
		Result: pb.GenProofResponse_RESULT_GEN_PROOF_OK,
	}, nil
}

func (zkp *zkProverServiceServer) GetProof(srv pb.ZKProverService_GetProofServer) error {
	newStateRoot, _ := new(big.Int).SetString("1212121212121212121212121212121212121212121212121212121212121212", 16)
	newLocalExitRoot, _ := new(big.Int).SetString("1234123412341234123412341234123412341234123412341234123412341234", 16)
	publicInputs := &pb.PublicInputs{
		NewStateRoot:     newStateRoot.String(),
		NewLocalExitRoot: newLocalExitRoot.String(),
	}

	ctx := srv.Context()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		req, err := srv.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			fmt.Printf("GetProof err: %v", err)
			continue
		}
		if st, ok := zkp.idsToState[req.Id]; ok {
			if st == 1 {
				resp := &pb.GetProofResponse{
					Id:     req.Id,
					Proof:  mockProof,
					Public: &pb.PublicInputsExtended{PublicInputs: publicInputs},
					Result: pb.GetProofResponse_RESULT_GET_PROOF_COMPLETED_OK,
				}
				err := srv.Send(resp)
				if err != nil {
					fmt.Printf("Get proof err: %v\n", err)
				}
			} else if st == 0 {
				resp := &pb.GetProofResponse{
					Id:     req.Id,
					Result: pb.GetProofResponse_RESULT_GET_PROOF_PENDING,
				}
				zkp.idsToState[req.Id] = 1
				err := srv.Send(resp)
				if err != nil {
					fmt.Printf("Get proof err: %v\n", err)
				}
			}
		} else {
			resp := &pb.GetProofResponse{
				Id:     req.Id,
				Result: pb.GetProofResponse_RESULT_GET_PROOF_ERROR,
			}
			err := srv.Send(resp)
			if err != nil {
				fmt.Printf("Get proof err: %v\n", err)
			}
		}
	}
}

func (zkp *zkProverServiceServer) GetStatus(ctx context.Context, request *pb.GetStatusRequest) (*pb.GetStatusResponse, error) {
	return &pb.GetStatusResponse{
		State:                     pb.GetStatusResponse_STATUS_PROVER_IDLE,
		LastComputedRequestId:     strconv.Itoa(zkp.id),
		LastComputedEndTime:       uint64(time.Now().Unix()),
		CurrentComputingRequestId: strconv.Itoa(zkp.id + 1),
		CurrentComputingStartTime: 0,
		VersionProto:              serverProtoVersion,
		VersionServer:             serverVersion,
		PendingRequestQueueIds:    []string{},
	}, nil
}

func (zkp *zkProverServiceServer) Cancel(ctx context.Context, request *pb.CancelRequest) (*pb.CancelResponse, error) {
	return &pb.CancelResponse{Result: pb.CancelResponse_RESULT_CANCEL_OK}, nil
}

func (zkp *zkProverServiceServer) Execute(server pb.ZKProverService_ExecuteServer) error {
	return nil
}

func (zkp *zkProverServiceServer) SynchronizeBatchProposal(ctx context.Context, request *pb.SynchronizeBatchProposalRequest) (*pb.SynchronizeBatchProposalResponse, error) {
	return nil, nil
}