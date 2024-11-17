package processor

import (
	"fmt"
	"strings"
)

// https://docs.aws.amazon.com/elasticloadbalancing/latest/application/load-balancer-access-logs.html#access-log-entry-format
var fieldNames = []string{
	"type",
	"time",
	"elb",
	"client:port",
	"target:port",
	"request_processing_time",
	"target_processing_time",
	"response_processing_time",
	"elb_status_code",
	"target_status_code",
	"received_bytes",
	"sent_bytes",
	"request",
	"user_agent",
	"ssl_cipher",   // https listener
	"ssl_protocol", // https listener
	"target_group_arn",
	"trace_id",
	"domain_name",     // https listener
	"chosen_cert_arn", // https listener
	"matched_rule_priority",
	"request_creation_time",
	"actions_executed",
	"redirect_url",
	"error_reason",
	"target:port_list",
	"target_status_code_list",
	"classification",
	"classification_reason",
	"conn_trace_id", // https listener
}

type Fields interface {
	GetFieldNameByIndex(index int) (string, error)
	IncludeField(index int) bool
}

type IncludedFields struct {
	includedFieldsMap map[string]bool
}

func NewFields(fieldsConfig string) (*IncludedFields, error) {
	fs := &IncludedFields{
		includedFieldsMap: make(map[string]bool),
	}
	var validFieldMap = make(map[string]bool)
	for _, field := range fieldNames {
		validFieldMap[field] = true
	}
	// If no fields are provided, include all fields:
	if fieldsConfig == "" {
		for field := range validFieldMap {
			fs.includedFieldsMap[field] = true
		}

		return fs, nil
	}
	// Include only the fields that are provided in config:
	fields := strings.Split(fieldsConfig, ",")
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if _, ok := validFieldMap[field]; !ok {
			return nil, fmt.Errorf("invalid field name '%s' provided", field)
		}
		fs.includedFieldsMap[field] = true
	}

	return fs, nil
}

func (fs *IncludedFields) GetFieldNameByIndex(index int) (string, error) {
	if index < 0 || index >= len(fieldNames) {
		return "", fmt.Errorf("invalid field index %d", index)
	}

	return fieldNames[index], nil
}

func (fs *IncludedFields) IncludeField(index int) bool {
	if index < 0 || index >= len(fieldNames) {
		return false
	}
	fieldName := fieldNames[index]
	exists := fs.includedFieldsMap[fieldName]

	return exists
}
