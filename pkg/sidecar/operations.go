package sidecar

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Orange-OpenSource/casskop/pkg/common/operations"
	"github.com/google/uuid"
)

//go:generate jsonenums -type=Kind
type Kind int

// Operations

const (
	noop Kind = iota
	backup
)

type operationsFilter struct {
	Types  []Kind
	States []operations.OperationState
}

func (o operationsFilter) buildFilteredEndpoint(endpoint string) string {
	var filterT, filterS string
	if len(o.Types) > 0 {
		var kinds []string
		for _, kind := range o.Types {
			kinds = append(kinds, _KindValueToName[kind])
		}
		filterT = "type=" + strings.Join(kinds, ",")
	}

	if len(o.States) > 0 {
		var states []string
		for _, state := range o.States {
			states = append(states, string(state))
		}
		filterS = "state=" + strings.Join(states, ",")
	}

	filter := strings.Join([]string{filterT, filterS}, "&")
	if len(filter) > 0 {
		return endpoint + "?" + filter
	}

	return endpoint

}

type operationRequest interface {
	Init()
}

type operation struct {
	Type Kind `json:"type"`
}

type BackupRequest struct {
	operation
	StorageLocation       string   `json:"storageLocation"`
	SnapshotTag           string   `json:"snapshotTag,omitempty"`
	Duration              string   `json:"duration,omitempty"`
	Bandwidth             string   `json:"bandwidth,omitempty"`
	ConcurrentConnections int      `json:"concurrentConnections,omitempty"`
	Entities              string   `json:"table,omitempty"`
	Keyspaces             []string `json:"keyspaces,omitempty"`
	Secret                string   `json:"k8sSecretName"`
	KubernetesNamespace   string   `json:"k8sNamespace"`
	GlobalRequest         bool     `json:"globalRequest,omitempty"`
	Datacenter            string   `json:"dc"`
}

func (b *BackupRequest) Init() {
	b.Type = backup
}

type tokenRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type operationResponse map[string]interface{}
type Operations []operationResponse

type basicResponse struct {
	Id             uuid.UUID                 `json:"id"`
	CreationTime   time.Time                 `json:"creationTime"`
	State          operations.OperationState `json:"state"`
	Progress       float32                   `json:"progress"`
	StartTime      time.Time                 `json:"startTime"`
	CompletionTime time.Time                 `json:"completionTime"`
}

func (client *Client) FindBackup(id uuid.UUID) (backupResponse *BackupResponse, err error) {
	if op, err := client.GetOperation(id); err != nil {
		return nil, err
	} else if b, err := ParseOperation(*op, backup); err != nil {
		return nil, err
	} else if backupResponse, ok := b.(*BackupResponse); !ok {
		return nil, fmt.Errorf("can't parse operation to backup")
	} else {
		return backupResponse, nil
	}
}

// backup Operations
type BackupResponse struct {
	basicResponse
	BackupRequest
}

func (b *BackupResponse) String() string {
	op, _ := json.Marshal(b)
	return string(op)
}

func (client *Client) ListBackups() ([]*BackupResponse, error) {

	ops, err := client.GetOperations()
	if ops == nil || err != nil {
		return []*BackupResponse{}, err
	}

	backupOps, err := FilterOperations(*ops, backup)
	if err != nil {
		return []*BackupResponse{}, err

	}

	var backups []*BackupResponse
	for _, op := range backupOps {
		backups = append(backups, op.(*BackupResponse))
	}

	return backups, nil
}