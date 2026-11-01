package envVars

import (
	"fmt"
	"math"

	"github.com/Jeffail/gabs"
)

type renamedKey struct {
	name string
	// unit implied by the legacy name, appended to numeric values (empty when the rename carries no unit)
	unit string
}

// legacyCassandraYamlKeys maps the cassandra.yaml settings renamed by Cassandra 4.1 (CASSANDRA-15234)
// and 5.0 to their new name. It mirrors the @Replaces annotations of
// org.apache.cassandra.config.Config in Cassandra 5.0, except for the keyspace/table count guardrails
// whose value is shifted by Cassandra itself.
var legacyCassandraYamlKeys = map[string]renamedKey{
	"batch_size_fail_threshold_in_kb":                                        {"batch_size_fail_threshold", "KiB"},
	"batch_size_warn_threshold_in_kb":                                        {"batch_size_warn_threshold", "KiB"},
	"batchlog_replay_throttle_in_kb":                                         {"batchlog_replay_throttle", "KiB"},
	"cache_load_timeout_seconds":                                             {"cache_load_timeout", "s"},
	"cas_contention_timeout_in_ms":                                           {"cas_contention_timeout", "ms"},
	"cdc_free_space_check_interval_ms":                                       {"cdc_free_space_check_interval", "ms"},
	"cdc_total_space_in_mb":                                                  {"cdc_total_space", "MiB"},
	"column_index_cache_size_in_kb":                                          {"column_index_cache_size", "KiB"},
	"column_index_size_in_kb":                                                {"column_index_size", "KiB"},
	"commitlog_segment_size_in_mb":                                           {"commitlog_segment_size", "MiB"},
	"commitlog_sync_group_window_in_ms":                                      {"commitlog_sync_group_window", "ms"},
	"commitlog_sync_period_in_ms":                                            {"commitlog_sync_period", "ms"},
	"commitlog_total_space_in_mb":                                            {"commitlog_total_space", "MiB"},
	"compaction_large_partition_warning_threshold_mb":                        {"partition_size_warn_threshold", "MiB"},
	"compaction_throughput_mb_per_sec":                                       {"compaction_throughput", "MiB/s"},
	"compaction_tombstone_warning_threshold":                                 {"partition_tombstones_warn_threshold", ""},
	"counter_cache_save_period":                                              {"counter_cache_save_period", "s"},
	"counter_cache_size_in_mb":                                               {"counter_cache_size", "MiB"},
	"counter_write_request_timeout_in_ms":                                    {"counter_write_request_timeout", "ms"},
	"credentials_update_interval_in_ms":                                      {"credentials_update_interval", "ms"},
	"credentials_validity_in_ms":                                             {"credentials_validity", "ms"},
	"cross_node_timeout":                                                     {"internode_timeout", ""},
	"dynamic_snitch_reset_interval_in_ms":                                    {"dynamic_snitch_reset_interval", "ms"},
	"dynamic_snitch_update_interval_in_ms":                                   {"dynamic_snitch_update_interval", "ms"},
	"enable_drop_compact_storage":                                            {"drop_compact_storage_enabled", ""},
	"enable_materialized_views":                                              {"materialized_views_enabled", ""},
	"enable_sasi_indexes":                                                    {"sasi_indexes_enabled", ""},
	"enable_scripted_user_defined_functions":                                 {"scripted_user_defined_functions_enabled", ""},
	"enable_transient_replication":                                           {"transient_replication_enabled", ""},
	"enable_user_defined_functions":                                          {"user_defined_functions_enabled", ""},
	"enable_user_defined_functions_threads":                                  {"user_defined_functions_threads_enabled", ""},
	"file_cache_size_in_mb":                                                  {"file_cache_size", "MiB"},
	"gc_log_threshold_in_ms":                                                 {"gc_log_threshold", "ms"},
	"gc_warn_threshold_in_ms":                                                {"gc_warn_threshold", "ms"},
	"hinted_handoff_throttle_in_kb":                                          {"hinted_handoff_throttle", "KiB"},
	"hints_flush_period_in_ms":                                               {"hints_flush_period", "ms"},
	"index_summary_capacity_in_mb":                                           {"index_summary_capacity", "MiB"},
	"index_summary_resize_interval_in_minutes":                               {"index_summary_resize_interval", "m"},
	"inter_dc_stream_throughput_outbound_megabits_per_sec":                   {"inter_dc_stream_throughput_outbound", "Mbps"},
	"internode_application_receive_queue_capacity_in_bytes":                  {"internode_application_receive_queue_capacity", "B"},
	"internode_application_receive_queue_reserve_endpoint_capacity_in_bytes": {"internode_application_receive_queue_reserve_endpoint_capacity", "B"},
	"internode_application_receive_queue_reserve_global_capacity_in_bytes":   {"internode_application_receive_queue_reserve_global_capacity", "B"},
	"internode_application_send_queue_capacity_in_bytes":                     {"internode_application_send_queue_capacity", "B"},
	"internode_application_send_queue_reserve_endpoint_capacity_in_bytes":    {"internode_application_send_queue_reserve_endpoint_capacity", "B"},
	"internode_application_send_queue_reserve_global_capacity_in_bytes":      {"internode_application_send_queue_reserve_global_capacity", "B"},
	"internode_max_message_size_in_bytes":                                    {"internode_max_message_size", "B"},
	"internode_recv_buff_size_in_bytes":                                      {"internode_socket_receive_buffer_size", "B"},
	"internode_send_buff_size_in_bytes":                                      {"internode_socket_send_buffer_size", "B"},
	"internode_socket_receive_buffer_size_in_bytes":                          {"internode_socket_receive_buffer_size", "B"},
	"internode_socket_send_buffer_size_in_bytes":                             {"internode_socket_send_buffer_size", "B"},
	"internode_streaming_tcp_user_timeout_in_ms":                             {"internode_streaming_tcp_user_timeout", "ms"},
	"internode_tcp_connect_timeout_in_ms":                                    {"internode_tcp_connect_timeout", "ms"},
	"internode_tcp_user_timeout_in_ms":                                       {"internode_tcp_user_timeout", "ms"},
	"key_cache_save_period":                                                  {"key_cache_save_period", "s"},
	"key_cache_size_in_mb":                                                   {"key_cache_size", "MiB"},
	"max_hint_window_in_ms":                                                  {"max_hint_window", "ms"},
	"max_hints_file_size_in_mb":                                              {"max_hints_file_size", "MiB"},
	"max_mutation_size_in_kb":                                                {"max_mutation_size", "KiB"},
	"max_value_size_in_mb":                                                   {"max_value_size", "MiB"},
	"memtable_heap_space_in_mb":                                              {"memtable_heap_space", "MiB"},
	"memtable_offheap_space_in_mb":                                           {"memtable_offheap_space", "MiB"},
	"min_free_space_per_drive_in_mb":                                         {"min_free_space_per_drive", "MiB"},
	"native_transport_idle_timeout_in_ms":                                    {"native_transport_idle_timeout", "ms"},
	"native_transport_max_concurrent_requests_in_bytes":                      {"native_transport_max_request_data_in_flight", "B"},
	"native_transport_max_concurrent_requests_in_bytes_per_ip":               {"native_transport_max_request_data_in_flight_per_ip", "B"},
	"native_transport_max_frame_size_in_mb":                                  {"native_transport_max_frame_size", "MiB"},
	"native_transport_receive_queue_capacity_in_bytes":                       {"native_transport_receive_queue_capacity", "B"},
	"networking_cache_size_in_mb":                                            {"networking_cache_size", "MiB"},
	"periodic_commitlog_sync_lag_block_in_ms":                                {"periodic_commitlog_sync_lag_block", "ms"},
	"permissions_update_interval_in_ms":                                      {"permissions_update_interval", "ms"},
	"permissions_validity_in_ms":                                             {"permissions_validity", "ms"},
	"prepared_statements_cache_size_mb":                                      {"prepared_statements_cache_size", "MiB"},
	"range_request_timeout_in_ms":                                            {"range_request_timeout", "ms"},
	"read_request_timeout_in_ms":                                             {"read_request_timeout", "ms"},
	"repair_request_timeout_in_ms":                                           {"repair_request_timeout", "ms"},
	"repair_session_space_in_mb":                                             {"repair_session_space", "MiB"},
	"request_timeout_in_ms":                                                  {"request_timeout", "ms"},
	"roles_update_interval_in_ms":                                            {"roles_update_interval", "ms"},
	"roles_validity_in_ms":                                                   {"roles_validity", "ms"},
	"row_cache_save_period":                                                  {"row_cache_save_period", "s"},
	"row_cache_size_in_mb":                                                   {"row_cache_size", "MiB"},
	"slow_query_log_timeout_in_ms":                                           {"slow_query_log_timeout", "ms"},
	"sstable_preemptive_open_interval_in_mb":                                 {"sstable_preemptive_open_interval", "MiB"},
	"stream_throughput_outbound_megabits_per_sec":                            {"stream_throughput_outbound", "Mbps"},
	"streaming_keep_alive_period_in_secs":                                    {"streaming_keep_alive_period", "s"},
	"tracetype_query_ttl":                                                    {"trace_type_query_ttl", "s"},
	"tracetype_repair_ttl":                                                   {"trace_type_repair_ttl", "s"},
	"trickle_fsync_interval_in_kb":                                           {"trickle_fsync_interval", "KiB"},
	"truncate_request_timeout_in_ms":                                         {"truncate_request_timeout", "ms"},
	"user_defined_function_fail_timeout":                                     {"user_defined_functions_fail_timeout", "ms"},
	"user_defined_function_warn_timeout":                                     {"user_defined_functions_warn_timeout", "ms"},
	"validation_preview_purge_head_start_in_sec":                             {"validation_preview_purge_head_start", "s"},
	"write_request_timeout_in_ms":                                            {"write_request_timeout", "ms"},
}

// translateLegacyCassandraYaml rewrites the legacy keys of a cassandra-yaml section into their
// Cassandra 4.1+ names so that casskop defaults and configurations written for older versions keep
// working and cannot clash with their new style counterpart. A new style key already present wins.
func translateLegacyCassandraYaml(section map[string]any) map[string]any {
	translated := make(map[string]any, len(section))
	for key, value := range section {
		renamed, legacy := legacyCassandraYamlKeys[key]
		if !legacy {
			translated[key] = value
			continue
		}
		if _, exists := section[renamed.name]; exists {
			continue
		}
		if value, ok := legacyValue(value, renamed.unit); ok {
			translated[renamed.name] = value
		}
	}
	return translated
}

// legacyValue converts a legacy value to its 4.1+ representation: numbers get the unit implied by the
// legacy name while strings are kept as is (they are assumed to already carry a unit). Negative values
// used to mean "disabled" and are now expressed by omitting the key, hence the false return.
func legacyValue(value any, unit string) (any, bool) {
	var number float64
	switch v := value.(type) {
	case int:
		number = float64(v)
	case int64:
		number = float64(v)
	case float64:
		number = v
	default:
		return value, true
	}
	if unit == "" {
		return value, true
	}
	if number < 0 {
		return nil, false
	}
	switch {
	case unit == "Mbps":
		return fmt.Sprintf("%dB/s", int64(number*125000)), true
	case number == math.Trunc(number):
		return fmt.Sprintf("%d%s", int64(number), unit), true
	default:
		return fmt.Sprintf("%v%s", number, unit), true
	}
}

// translateLegacyCassandraYamlSection applies translateLegacyCassandraYaml to the cassandra-yaml
// section of a parsed config builder configuration
func translateLegacyCassandraYamlSection(parsedConfig *gabs.Container) {
	if section, ok := parsedConfig.Path(cassandraYamlKey).Data().(map[string]any); ok {
		// cassandraYamlKey is a constant top-level path, the call cannot fail
		_, _ = parsedConfig.SetP(translateLegacyCassandraYaml(section), cassandraYamlKey)
	}
}
