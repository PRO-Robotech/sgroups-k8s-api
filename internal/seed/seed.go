package seed

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"google.golang.org/protobuf/encoding/protojson"

	"sgroups.io/sgroups-k8s-api/internal/backend"

	sgroupsv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
)

type seedFile struct {
	Namespaces    json.RawMessage `json:"namespaces"`
	AddressGroups json.RawMessage `json:"addressGroups"`
}

func Load(path string, namespaces backend.NamespaceBackend, addressGroups backend.AddressGroupBackend) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var seed seedFile
	if err := json.Unmarshal(data, &seed); err != nil {
		return fmt.Errorf("seed parse failed: %w", err)
	}

	ctx := context.Background()
	if len(seed.Namespaces) > 0 {
		if namespaces == nil {
			return errors.New("namespace backend is nil")
		}

		wrapped := append([]byte(`{"namespaces":`), seed.Namespaces...)
		wrapped = append(wrapped, '}')
		var list sgroupsv1.NamespaceReq_Upsert
		if err := protojson.Unmarshal(wrapped, &list); err != nil {
			return fmt.Errorf("seed namespaces decode failed: %w", err)
		}
		if len(list.GetNamespaces()) > 0 {
			if _, err := namespaces.UpsertNamespaces(ctx, &list); err != nil {
				return fmt.Errorf("seed namespaces upsert failed: %w", err)
			}
		}
	}

	if len(seed.AddressGroups) > 0 {
		if addressGroups == nil {
			return errors.New("address group backend is nil")
		}

		wrapped := append([]byte(`{"addressGroups":`), seed.AddressGroups...)
		wrapped = append(wrapped, '}')
		var list sgroupsv1.AddressGroupReq_Upsert
		if err := protojson.Unmarshal(wrapped, &list); err != nil {
			return fmt.Errorf("seed addressGroups decode failed: %w", err)
		}
		if len(list.GetAddressGroups()) > 0 {
			if _, err := addressGroups.UpsertAddressGroups(ctx, &list); err != nil {
				return fmt.Errorf("seed addressGroups upsert failed: %w", err)
			}
		}
	}

	return nil
}
