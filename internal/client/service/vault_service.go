package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/msorokin-hash/passkeeper/internal/entity"
	proto "github.com/msorokin-hash/passkeeper/internal/protobuf"
	"google.golang.org/grpc/status"
)

// AddBankCard adds a bank card data to the Vault service.
//
// If metadata serialization fails, the method returns an error without
// attempting a gRPC call.
// If the gRPC request fails, the error is returned in a human-readable form.
func (s *Service) AddBankCard(ctx context.Context, data entity.BankCardData, meta entity.BankCardMeta) error {
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("failed to marshal bank card metadata: %w", err)
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal bank card data: %w", err)
	}
	return s.add(ctx, dataBytes, string(metaBytes), entity.BankCard)
}

// AddFile adds a file to the Vault service with associated metadata.
//
// The method validates the file path, reads file contents, and serializes
// metadata before making a gRPC call. If any validation or preparation step
// fails, the method returns an error without attempting the gRPC call.
// If the gRPC request fails, the error is returned in a human-readable form.
//
// Returns:
//   - error: nil on success, or descriptive error on failure including:
//   - file not found or inaccessible
//   - metadata serialization failure
//   - gRPC communication failure
func (s *Service) AddFile(ctx context.Context, file string, comment string) error {
	fileInfo, err := os.Stat(file)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %w", err)
		}
		return fmt.Errorf("invalid or inaccessible file path: %w", err)
	}

	dataBytes, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	meta := entity.FileMeta{
		Name:      fileInfo.Name(),
		Extension: filepath.Ext(file),
		Comment:   comment,
	}

	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("failed to marshal file metadata: %w", err)
	}

	return s.add(ctx, dataBytes, string(metaBytes), entity.File)
}

// AddPassword adds a password record to the Vault service.
//
// If metadata serialization fails, the method returns an error without
// attempting a gRPC call.
// If the gRPC request fails, the error is returned in a human-readable form.
func (s *Service) AddPassword(ctx context.Context, data string, meta entity.PasswordMeta) error {
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("failed to marshal password metadata: %w", err)
	}
	return s.add(ctx, []byte(data), string(metaBytes), entity.Password)
}

// AddText adds a text record to the Vault service.
//
// If metadata serialization fails, the method returns an error without
// attempting a gRPC call.
// If the gRPC request fails, the error is returned in a human-readable form.
func (s *Service) AddText(ctx context.Context, data string, meta entity.TextMeta) error {
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("failed to marshal text metadata: %w", err)
	}
	return s.add(ctx, []byte(data), string(metaBytes), entity.Text)
}

// DeleteData removes a stored data record from the Vault service by its ID.
// It sends the provided data ID to VaultService.DeleteData.
//
// It returns nil if request was succeed or an error.
func (s *Service) DeleteData(ctx context.Context, id string) error {
	_, err := s.grpcClient.VaultClient.DeleteData(ctx, &proto.DeleteDataRequest{
		Id: id,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			return fmt.Errorf("failed to delete data: %s", st.Message())
		}
		return err
	}
	return nil
}

// GetAllByType fetches all items from the Vault service by its data type.
// It sends the provided data type to VaultService.GetAllByType.
//
// It returns a slice of ItemInfo with deserialized metadata or an error.
func (s *Service) GetAllByType(ctx context.Context, dataType entity.DataType) ([]entity.ItemInfo, error) {
	res, err := s.grpcClient.VaultClient.GetAllByType(ctx, &proto.GetAllByTypeRequest{
		RecordType: string(dataType),
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			return nil, fmt.Errorf("failed to get data: %s", st.Message())
		}
		return nil, err
	}

	items := res.GetItems()
	result := make([]entity.ItemInfo, 0, len(items))

	newMeta := func(dt entity.DataType) (any, error) {
		switch dt {
		case entity.Password:
			return &entity.PasswordMeta{}, nil
		case entity.BankCard:
			return &entity.BankCardMeta{}, nil
		case entity.Text:
			return &entity.TextMeta{}, nil
		case entity.File:
			return &entity.FileMeta{}, nil
		default:
			return nil, fmt.Errorf("unknown data type: %s", dt)
		}
	}

	for _, item := range items {
		meta, err := newMeta(dataType)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(item.GetMeta()), meta); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata for item %q: %w", item.GetId(), err)
		}

		result = append(result, entity.ItemInfo{
			ID:   item.GetId(),
			Meta: meta,
		})
	}

	return result, nil
}

// GetData fetches a single data record from the Vault service by its ID.
// It calls VaultService.GetData and deserializes the returned metadata and data
// into the corresponding domain structures based on the record type.
//
// It returns:
//   - the resolved entity.DataType,
//   - a concrete item struct (e.g. *entity.PasswordItem, *entity.BankCardItem, etc.) as `any`,
//   - or a non-nil error on failure.
func (s *Service) GetData(ctx context.Context, id string) (entity.DataType, any, error) {
	res, err := s.grpcClient.VaultClient.GetData(ctx, &proto.GetDataRequest{
		Id: id,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			return "", nil, fmt.Errorf("failed to get data: %s", st.Message())
		}
		return "", nil, fmt.Errorf("failed to get data: %w", err)
	}

	resItem := res.GetItem()
	if resItem == nil {
		return "", nil, fmt.Errorf("empty item in server response")
	}

	itemType := resItem.GetRecordType()
	metaBytes := []byte(resItem.GetMeta())
	dataBytes := resItem.GetData()

	switch itemType {
	case string(entity.Password):
		meta := &entity.PasswordMeta{}
		if err := json.Unmarshal(metaBytes, meta); err != nil {
			return "", nil, fmt.Errorf("failed to decode metadata: %w", err)
		}
		return entity.Password, &entity.PasswordItem{
			Type: entity.Password,
			Meta: *meta,
			Data: string(dataBytes),
		}, nil

	case string(entity.BankCard):
		meta := &entity.BankCardMeta{}
		if err := json.Unmarshal(metaBytes, meta); err != nil {
			return "", nil, fmt.Errorf("failed to decode metadata: %w", err)
		}
		data := &entity.BankCardData{}
		if err := json.Unmarshal(dataBytes, data); err != nil {
			return "", nil, fmt.Errorf("failed to decode data: %w", err)
		}
		return entity.BankCard, &entity.BankCardItem{
			Type: entity.BankCard,
			Meta: *meta,
			Data: entity.BankCardData{
				Number:     data.Number,
				Holder:     data.Holder,
				CSV:        data.CSV,
				ValidMonth: data.ValidMonth,
				ValidYear:  data.ValidYear,
			},
		}, nil

	case string(entity.Text):
		meta := &entity.TextMeta{}
		if err := json.Unmarshal(metaBytes, meta); err != nil {
			return "", nil, fmt.Errorf("failed to decode metadata: %w", err)
		}
		return entity.Text, &entity.TextItem{
			Type: entity.Text,
			Meta: *meta,
			Data: string(dataBytes),
		}, nil

	case string(entity.File):
		meta := &entity.FileMeta{}
		if err := json.Unmarshal(metaBytes, meta); err != nil {
			return "", nil, fmt.Errorf("failed to decode metadata: %w", err)
		}

		filename := filepath.Join(s.workDir, meta.Name)

		if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
			return "", nil, fmt.Errorf("failed to remove existing file %s: %w", filename, err)
		}

		if err := os.WriteFile(filename, dataBytes, 0o644); err != nil {
			return "", nil, fmt.Errorf("failed to save file %s: %w", filename, err)
		}

		return entity.File, &entity.FileItem{
			Type: entity.File,
			Meta: *meta,
		}, nil
	}

	return "", nil, fmt.Errorf("unknown data type: %s", itemType)
}

// add sends data to the Vault service via gRPC.
// Returns:
//   - error: Returns nil on success. On failure, returns either a formatted error
//     containing the gRPC status message (if available) or the raw gRPC error.
func (s *Service) add(ctx context.Context, data []byte, meta string, dataType entity.DataType) error {
	_, err := s.grpcClient.VaultClient.AddData(ctx, &proto.AddDataRequest{
		Item: &proto.Item{
			Data:       data,
			RecordType: string(dataType),
			Meta:       meta,
		},
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			return fmt.Errorf("failed to save data: %s", st.Message())
		}
		return fmt.Errorf("failed to save data: %w", err)
	}
	return nil
}
