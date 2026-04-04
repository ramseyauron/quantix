// MIT License
// Copyright (c) 2024 quantix

// go/src/p2p/consensus_nodemanager.go
package p2p

// P2PNodeManager implements consensus.NodeManager using the TCP transport layer.
// This enables cross-process PBFT consensus by serializing consensus messages
// (proposal, vote, prepare, timeout) as security.Message packets and broadcasting
// them over all active TCP connections via transport.BroadcastToAll.
//
// This replaces the in-memory CallNodeManager for multi-node testnet scenarios.

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/ramseyauron/quantix/src/consensus"
	security "github.com/ramseyauron/quantix/src/handshake"
	"github.com/ramseyauron/quantix/src/network"
	"github.com/ramseyauron/quantix/src/transport"
)

// P2PNodeManager implements consensus.NodeManager via P2P TCP broadcast.
type P2PNodeManager struct {
	mu          sync.RWMutex
	nodeManager *network.NodeManager // underlying network node manager
}

// NewP2PNodeManager creates a new P2P-backed consensus node manager.
func NewP2PNodeManager(nm *network.NodeManager) *P2PNodeManager {
	return &P2PNodeManager{nodeManager: nm}
}

// GetPeers returns a map of peer IDs to consensus.Peer.
func (m *P2PNodeManager) GetPeers() map[string]consensus.Peer {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make(map[string]consensus.Peer)
	if m.nodeManager == nil {
		return result
	}
	for id, p := range m.nodeManager.GetPeers() {
		result[id] = &p2pPeer{peer: *p}
	}
	return result
}

// GetNode returns a consensus.Node for the given node ID.
func (m *P2PNodeManager) GetNode(nodeID string) consensus.Node {
	if m.nodeManager == nil {
		return &p2pNode{id: nodeID}
	}
	n := m.nodeManager.GetNode(nodeID)
	if n == nil {
		return &p2pNode{id: nodeID}
	}
	return &p2pNode{id: n.ID}
}

// BroadcastMessage encodes a consensus message and broadcasts it via TCP to all peers.
func (m *P2PNodeManager) BroadcastMessage(messageType string, data interface{}) error {
	// JSON-encode the payload
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("P2PNodeManager: failed to encode %s: %w", messageType, err)
	}

	// Wrap as a consensus_msg envelope
	env := consensusMsgEnvelope{
		Type:    messageType,
		Payload: payload,
	}
	envBytes, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("P2PNodeManager: failed to encode envelope: %w", err)
	}

	msg := &security.Message{
		Type: "consensus_msg",
		Data: string(envBytes),
	}

	errs := transport.BroadcastToAll(msg)
	for _, e := range errs {
		log.Printf("P2PNodeManager: BroadcastMessage(%s) error: %v", messageType, e)
	}
	log.Printf("P2PNodeManager: broadcast %s to all peers (%d errors)", messageType, len(errs))

	// Also deliver locally (this node also needs to process the message)
	// This is handled by consensusLoop directly; no self-delivery needed here.
	return nil
}

// BroadcastRANDAOState broadcasts RANDAO state to peers.
// For now this is a no-op for cross-process P2P — in-memory RANDAO sync
// suffices for single-host testnet; full cross-process sync is future work.
func (m *P2PNodeManager) BroadcastRANDAOState(mix [32]byte, submissions map[uint64]map[string]*consensus.VDFSubmission) error {
	// TODO: encode and broadcast RANDAO state over TCP
	return nil
}

// consensusMsgEnvelope is the wire format for cross-process consensus messages.
type consensusMsgEnvelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// DecodeConsensusEnvelope decodes a raw envelope from the P2P layer.
func DecodeConsensusEnvelope(raw []byte) (*consensusMsgEnvelope, error) {
	var env consensusMsgEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, err
	}
	return &env, nil
}

// RouteConsensusMessage decodes and routes a raw consensus_msg to the given consensus engine.
func RouteConsensusMessage(raw []byte, cons *consensus.Consensus) error {
	env, err := DecodeConsensusEnvelope(raw)
	if err != nil {
		return fmt.Errorf("RouteConsensusMessage: decode envelope: %w", err)
	}

	switch env.Type {
	case "proposal":
		var proposal consensus.Proposal
		if err := json.Unmarshal(env.Payload, &proposal); err != nil {
			return fmt.Errorf("RouteConsensusMessage: decode proposal: %w", err)
		}
		log.Printf("🔵 [P2P] Received proposal from %s", proposal.ProposerID)
		return cons.HandleProposal(&proposal)

	case "vote":
		var vote consensus.Vote
		if err := json.Unmarshal(env.Payload, &vote); err != nil {
			return fmt.Errorf("RouteConsensusMessage: decode vote: %w", err)
		}
		log.Printf("🟢 [P2P] Received CommitVote from %s", vote.VoterID)
		return cons.HandleVote(&vote)

	case "prepare":
		var vote consensus.Vote
		if err := json.Unmarshal(env.Payload, &vote); err != nil {
			return fmt.Errorf("RouteConsensusMessage: decode prepare vote: %w", err)
		}
		log.Printf("🟡 [P2P] Received PrepareVote from %s", vote.VoterID)
		return cons.HandlePrepareVote(&vote)

	case "timeout":
		var timeout consensus.TimeoutMsg
		if err := json.Unmarshal(env.Payload, &timeout); err != nil {
			return fmt.Errorf("RouteConsensusMessage: decode timeout: %w", err)
		}
		log.Printf("⏰ [P2P] Received Timeout from %s", timeout.VoterID)
		return cons.HandleTimeout(&timeout)

	default:
		return fmt.Errorf("RouteConsensusMessage: unknown type %q", env.Type)
	}
}

// --- consensus.Node implementation ---

type p2pNode struct {
	id string
}

func (n *p2pNode) GetID() string                   { return n.id }
func (n *p2pNode) GetRole() consensus.NodeRole     { return consensus.NodeRole(0) }
func (n *p2pNode) GetStatus() consensus.NodeStatus { return consensus.NodeStatus(0) }

// --- consensus.Peer implementation ---

type p2pPeer struct {
	peer network.Peer
}

func (p *p2pPeer) GetNode() consensus.Node {
	n := p.peer.Node
	if n == nil {
		return &p2pNode{}
	}
	return &p2pNode{id: n.ID}
}
