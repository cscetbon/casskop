package envVars

import (
	"math"
	"strconv"

	"github.com/Jeffail/gabs"
)

type renamedKey struct {
	name string
	// unit implied by the legacy name, appended to numeric values (empty when the rename carries no unit)
	unit string
	// negative is what a negative legacy value becomes, as done by the converters of Cassandra
	negative negativeValue
}

// negativeValue tells how a negative legacy value, which used to mean "disabled", is expressed in 4.1+
type negativeValue int

const (
	// negativeOmitted drops the setting, for the ones whose default is already null
	negativeOmitted negativeValue = iota
	// negativeNull sets an explicit null, for the settings enabled by default
	negativeNull
	// negativeZero sets 0, for the settings that no longer accept null
	negativeZero
)

// legacyCassandraYamlKeys maps the cassandra.yaml settings renamed by Cassandra 4.1 (CASSANDRA-15234)
// to their new name. It mirrors the @Replaces annotations of org.apache.cassandra.config.Config,
// except for the keyspace/table count guardrails whose value is shifted by Cassandra itself.
var legacyCassandraYamlKeys = map[string]renamedKey{
	"batch_size_fail_threshold_in_kb":                                        {"batch_size_fail_threshold", "KiB", negativeOmitted},
	"batch_size_warn_threshold_in_kb":                                        {"batch_size_warn_threshold", "KiB", negativeOmitted},
	"batchlog_replay_throttle_in_kb":                                         {"batchlog_replay_throttle", "KiB", negativeOmitted},
	"cache_load_timeout_seconds":                                             {"cache_load_timeout", "s", negativeZero},
	"cas_contention_timeout_in_ms":                                           {"cas_contention_timeout", "ms", negativeOmitted},
	"cdc_free_space_check_interval_ms":                                       {"cdc_free_space_check_interval", "ms", negativeOmitted},
	"cdc_total_space_in_mb":                                                  {"cdc_total_space", "MiB", negativeOmitted},
	"column_index_cache_size_in_kb":                                          {"column_index_cache_size", "KiB", negativeOmitted},
	"column_index_size_in_kb":                                                {"column_index_size", "KiB", negativeOmitted},
	"commitlog_segment_size_in_mb":                                           {"commitlog_segment_size", "MiB", negativeOmitted},
	"commitlog_sync_group_window_in_ms":                                      {"commitlog_sync_group_window", "ms", negativeOmitted},
	"commitlog_sync_period_in_ms":                                            {"commitlog_sync_period", "ms", negativeOmitted},
	"commitlog_total_space_in_mb":                                            {"commitlog_total_space", "MiB", negativeOmitted},
	"compaction_large_partition_warning_threshold_mb":                        {"compaction_large_partition_warning_threshold", "MiB", negativeOmitted},
	"compaction_throughput_mb_per_sec":                                       {"compaction_throughput", "MiB/s", negativeOmitted},
	"counter_cache_save_period":                                              {"counter_cache_save_period", "s", negativeOmitted},
	"counter_cache_size_in_mb":                                               {"counter_cache_size", "MiB", negativeOmitted},
	"counter_write_request_timeout_in_ms":                                    {"counter_write_request_timeout", "ms", negativeOmitted},
	"credentials_update_interval_in_ms":                                      {"credentials_update_interval", "ms", negativeOmitted},
	"credentials_validity_in_ms":                                             {"credentials_validity", "ms", negativeOmitted},
	"cross_node_timeout":                                                     {"internode_timeout", "", negativeOmitted},
	"dynamic_snitch_reset_interval_in_ms":                                    {"dynamic_snitch_reset_interval", "ms", negativeOmitted},
	"dynamic_snitch_update_interval_in_ms":                                   {"dynamic_snitch_update_interval", "ms", negativeOmitted},
	"enable_drop_compact_storage":                                            {"drop_compact_storage_enabled", "", negativeOmitted},
	"enable_materialized_views":                                              {"materialized_views_enabled", "", negativeOmitted},
	"enable_sasi_indexes":                                                    {"sasi_indexes_enabled", "", negativeOmitted},
	"enable_scripted_user_defined_functions":                                 {"scripted_user_defined_functions_enabled", "", negativeOmitted},
	"enable_transient_replication":                                           {"transient_replication_enabled", "", negativeOmitted},
	"enable_user_defined_functions":                                          {"user_defined_functions_enabled", "", negativeOmitted},
	"enable_user_defined_functions_threads":                                  {"user_defined_functions_threads_enabled", "", negativeOmitted},
	"file_cache_size_in_mb":                                                  {"file_cache_size", "MiB", negativeOmitted},
	"gc_log_threshold_in_ms":                                                 {"gc_log_threshold", "ms", negativeOmitted},
	"gc_warn_threshold_in_ms":                                                {"gc_warn_threshold", "ms", negativeOmitted},
	"hinted_handoff_throttle_in_kb":                                          {"hinted_handoff_throttle", "KiB", negativeOmitted},
	"hints_flush_period_in_ms":                                               {"hints_flush_period", "ms", negativeOmitted},
	"index_summary_capacity_in_mb":                                           {"index_summary_capacity", "MiB", negativeOmitted},
	"index_summary_resize_interval_in_minutes":                               {"index_summary_resize_interval", "m", negativeNull},
	"inter_dc_stream_throughput_outbound_megabits_per_sec":                   {"inter_dc_stream_throughput_outbound", "Mbps", negativeOmitted},
	"internode_application_receive_queue_capacity_in_bytes":                  {"internode_application_receive_queue_capacity", "B", negativeOmitted},
	"internode_application_receive_queue_reserve_endpoint_capacity_in_bytes": {"internode_application_receive_queue_reserve_endpoint_capacity", "B", negativeOmitted},
	"internode_application_receive_queue_reserve_global_capacity_in_bytes":   {"internode_application_receive_queue_reserve_global_capacity", "B", negativeOmitted},
	"internode_application_send_queue_capacity_in_bytes":                     {"internode_application_send_queue_capacity", "B", negativeOmitted},
	"internode_application_send_queue_reserve_endpoint_capacity_in_bytes":    {"internode_application_send_queue_reserve_endpoint_capacity", "B", negativeOmitted},
	"internode_application_send_queue_reserve_global_capacity_in_bytes":      {"internode_application_send_queue_reserve_global_capacity", "B", negativeOmitted},
	"internode_max_message_size_in_bytes":                                    {"internode_max_message_size", "B", negativeOmitted},
	"internode_recv_buff_size_in_bytes":                                      {"internode_socket_receive_buffer_size", "B", negativeOmitted},
	"internode_send_buff_size_in_bytes":                                      {"internode_socket_send_buffer_size", "B", negativeOmitted},
	"internode_socket_receive_buffer_size_in_bytes":                          {"internode_socket_receive_buffer_size", "B", negativeOmitted},
	"internode_socket_send_buffer_size_in_bytes":                             {"internode_socket_send_buffer_size", "B", negativeOmitted},
	"internode_streaming_tcp_user_timeout_in_ms":                             {"internode_streaming_tcp_user_timeout", "ms", negativeOmitted},
	"internode_tcp_connect_timeout_in_ms":                                    {"internode_tcp_connect_timeout", "ms", negativeOmitted},
	"internode_tcp_user_timeout_in_ms":                                       {"internode_tcp_user_timeout", "ms", negativeOmitted},
	"key_cache_save_period":                                                  {"key_cache_save_period", "s", negativeOmitted},
	"key_cache_size_in_mb":                                                   {"key_cache_size", "MiB", negativeOmitted},
	"max_hint_window_in_ms":                                                  {"max_hint_window", "ms", negativeOmitted},
	"max_hints_file_size_in_mb":                                              {"max_hints_file_size", "MiB", negativeOmitted},
	"max_mutation_size_in_kb":                                                {"max_mutation_size", "KiB", negativeOmitted},
	"max_value_size_in_mb":                                                   {"max_value_size", "MiB", negativeOmitted},
	"memtable_heap_space_in_mb":                                              {"memtable_heap_space", "MiB", negativeOmitted},
	"memtable_offheap_space_in_mb":                                           {"memtable_offheap_space", "MiB", negativeOmitted},
	"min_free_space_per_drive_in_mb":                                         {"min_free_space_per_drive", "MiB", negativeOmitted},
	"native_transport_idle_timeout_in_ms":                                    {"native_transport_idle_timeout", "ms", negativeOmitted},
	"native_transport_max_concurrent_requests_in_bytes":                      {"native_transport_max_request_data_in_flight", "B", negativeOmitted},
	"native_transport_max_concurrent_requests_in_bytes_per_ip":               {"native_transport_max_request_data_in_flight_per_ip", "B", negativeOmitted},
	"native_transport_max_frame_size_in_mb":                                  {"native_transport_max_frame_size", "MiB", negativeOmitted},
	"native_transport_receive_queue_capacity_in_bytes":                       {"native_transport_receive_queue_capacity", "B", negativeOmitted},
	"networking_cache_size_in_mb":                                            {"networking_cache_size", "MiB", negativeOmitted},
	"periodic_commitlog_sync_lag_block_in_ms":                                {"periodic_commitlog_sync_lag_block", "ms", negativeOmitted},
	"permissions_update_interval_in_ms":                                      {"permissions_update_interval", "ms", negativeOmitted},
	"permissions_validity_in_ms":                                             {"permissions_validity", "ms", negativeOmitted},
	"prepared_statements_cache_size_mb":                                      {"prepared_statements_cache_size", "MiB", negativeOmitted},
	"range_request_timeout_in_ms":                                            {"range_request_timeout", "ms", negativeOmitted},
	"read_request_timeout_in_ms":                                             {"read_request_timeout", "ms", negativeOmitted},
	"repair_request_timeout_in_ms":                                           {"repair_request_timeout", "ms", negativeOmitted},
	"repair_session_space_in_mb":                                             {"repair_session_space", "MiB", negativeOmitted},
	"request_timeout_in_ms":                                                  {"request_timeout", "ms", negativeOmitted},
	"roles_update_interval_in_ms":                                            {"roles_update_interval", "ms", negativeOmitted},
	"roles_validity_in_ms":                                                   {"roles_validity", "ms", negativeOmitted},
	"row_cache_save_period":                                                  {"row_cache_save_period", "s", negativeOmitted},
	"row_cache_size_in_mb":                                                   {"row_cache_size", "MiB", negativeOmitted},
	"slow_query_log_timeout_in_ms":                                           {"slow_query_log_timeout", "ms", negativeOmitted},
	"sstable_preemptive_open_interval_in_mb":                                 {"sstable_preemptive_open_interval", "MiB", negativeNull},
	"stream_throughput_outbound_megabits_per_sec":                            {"stream_throughput_outbound", "Mbps", negativeOmitted},
	"streaming_keep_alive_period_in_secs":                                    {"streaming_keep_alive_period", "s", negativeOmitted},
	"tracetype_query_ttl":                                                    {"trace_type_query_ttl", "s", negativeOmitted},
	"tracetype_repair_ttl":                                                   {"trace_type_repair_ttl", "s", negativeOmitted},
	"trickle_fsync_interval_in_kb":                                           {"trickle_fsync_interval", "KiB", negativeOmitted},
	"truncate_request_timeout_in_ms":                                         {"truncate_request_timeout", "ms", negativeOmitted},
	"user_defined_function_fail_timeout":                                     {"user_defined_functions_fail_timeout", "ms", negativeOmitted},
	"user_defined_function_warn_timeout":                                     {"user_defined_functions_warn_timeout", "ms", negativeOmitted},
	"validation_preview_purge_head_start_in_sec":                             {"validation_preview_purge_head_start", "s", negativeZero},
	"write_request_timeout_in_ms":                                            {"write_request_timeout", "ms", negativeOmitted},
}

// legacyCassandraYamlKeys50 holds the settings renamed again by Cassandra 5.0, which does not exist
// under that name in 4.1
var legacyCassandraYamlKeys50 = map[string]renamedKey{
	"compaction_large_partition_warning_threshold_mb": {"partition_size_warn_threshold", "MiB", negativeOmitted},
	"compaction_large_partition_warning_threshold":    {"partition_size_warn_threshold", "", negativeOmitted},
	"compaction_tombstone_warning_threshold":          {"partition_tombstones_warn_threshold", "", negativeOmitted},
}

// renamedCassandraYamlKey returns the name a legacy cassandra.yaml setting has in Cassandra 4.1, or in
// Cassandra 5.0 when cassandra5 is set
func renamedCassandraYamlKey(key string, cassandra5 bool) (renamedKey, bool) {
	if renamed, found := legacyCassandraYamlKeys50[key]; found && cassandra5 {
		return renamed, true
	}
	renamed, found := legacyCassandraYamlKeys[key]
	return renamed, found
}

// translateLegacyCassandraYaml rewrites the legacy keys of a cassandra-yaml section into their
// Cassandra 4.1 names, or 5.0 ones when cassandra5 is set, so that casskop defaults and configurations
// written for older versions keep working and cannot clash with their new style counterpart. A new
// style key already present wins.
func translateLegacyCassandraYaml(section map[string]any, cassandra5 bool) map[string]any {
	translated := make(map[string]any, len(section))
	for key, value := range section {
		renamed, legacy := renamedCassandraYamlKey(key, cassandra5)
		if !legacy {
			translated[key] = value
			continue
		}
		if _, exists := section[renamed.name]; exists {
			continue
		}
		if value, ok := legacyValue(value, renamed); ok {
			translated[renamed.name] = value
		}
	}
	return translated
}

// legacyValue converts a legacy value to its 4.1+ representation: numbers get the unit implied by the
// legacy name while strings are kept as is (they are assumed to already carry a unit). Negative values
// used to mean "disabled", which renamed.negative tells how to express; false is returned when the
// setting must be left out.
func legacyValue(value any, renamed renamedKey) (any, bool) {
	var number float64
	switch v := value.(type) {
	case int:
		number = float64(v)
	case float64:
		number = v
	default:
		return value, true
	}
	unit := renamed.unit
	switch {
	case unit == "":
		return value, true
	case number < 0 && renamed.negative == negativeOmitted:
		return nil, false
	case number < 0 && renamed.negative == negativeNull:
		return nil, true
	case number < 0:
		number = 0
	case unit == "Mbps":
		number, unit = math.Trunc(number*125000), "B/s"
	}
	return strconv.FormatFloat(number, 'f', -1, 64) + unit, true
}

// translateLegacyCassandraYamlSection applies translateLegacyCassandraYaml to the cassandra-yaml
// section of a parsed config builder configuration
func translateLegacyCassandraYamlSection(parsedConfig *gabs.Container, cassandra5 bool) {
	if section, ok := parsedConfig.Path(cassandraYamlKey).Data().(map[string]any); ok {
		// cassandraYamlKey is a constant top-level path, the call cannot fail
		_, _ = parsedConfig.SetP(translateLegacyCassandraYaml(section, cassandra5), cassandraYamlKey)
	}
}
