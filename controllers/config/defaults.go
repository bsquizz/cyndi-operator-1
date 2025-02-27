package config

const defaultTopic = "platform.inventory.events"
const defaultConnectCluster = "xjoin-kafka-connect-strimzi"
const defaultInventoryDbSecret = "host-inventory-db"

const defaultConnectorTemplate = `{
	"connector.class": "io.confluent.connect.jdbc.JdbcSinkConnector",
	"tasks.max": "{{.TasksMax}}",
	"topics": "{{.Topic}}",
	"key.converter": "org.apache.kafka.connect.storage.StringConverter",
	"value.converter": "org.apache.kafka.connect.json.JsonConverter",
	"value.converter.schemas.enable": false,
	"connection.url": "jdbc:postgresql://{{.DBHostname}}:{{.DBPort}}/{{.DBName}}",
	"connection.user": "{{.DBUser}}",
	"connection.password": "{{.DBPassword}}",
	"dialect.name": "EnhancedPostgreSqlDatabaseDialect",
	"auto.create": false,
	"insert.mode": "upsert",
	"delete.enabled": true,
	"batch.size": "{{.BatchSize}}",
	"table.name.format": "inventory.{{.TableName}}",
	"pk.mode": "record_key",
	"pk.fields": "id",
	"fields.whitelist": "account,display_name,tags,updated,created,stale_timestamp,system_profile,insights_id",

	{{ if eq .InsightsOnly "true" }}
	"transforms": "timestampFilter,insightsFilter,deleteToTombstone,extractHost,systemProfileFilter,systemProfileToJson,tagsToJson,injectSchemaKey,injectSchemaValue",
	"transforms.insightsFilter.type":"com.redhat.insights.kafka.connect.transforms.Filter",
	"transforms.insightsFilter.predicate": "!!record.headers().lastWithName('insights_id').value()",
	{{ else  }}
	"transforms": "timestampFilter,deleteToTombstone,extractHost,systemProfileFilter,systemProfileToJson,tagsToJson,injectSchemaKey,injectSchemaValue",
	{{ end }}

	"transforms.timestampFilter.type":"com.redhat.insights.kafka.connect.transforms.Filter",
	"transforms.timestampFilter.predicate": "(Date.now() - record.timestamp()) < {{.MaxAge}} * 24 * 60 * 60 * 1000",
	"transforms.deleteToTombstone.type":"com.redhat.insights.kafka.connect.transforms.DropIf$Value",
	"transforms.deleteToTombstone.predicate": "'delete'.equals(record.headers().lastWithName('event_type').value())",
	"transforms.extractHost.type":"org.apache.kafka.connect.transforms.ExtractField$Value",
	"transforms.extractHost.field":"host",
	"transforms.systemProfileFilter.type": "com.redhat.insights.kafka.connect.transforms.FilterFields$Value",
	"transforms.systemProfileFilter.field": "system_profile",
	"transforms.systemProfileFilter.allowlist": "{{.AllowlistSP}}",
	"transforms.systemProfileToJson.type": "com.redhat.insights.kafka.connect.transforms.FieldToJson$Value",
	"transforms.systemProfileToJson.originalField": "system_profile",
	"transforms.systemProfileToJson.destinationField": "system_profile",
	"transforms.tagsToJson.type": "com.redhat.insights.kafka.connect.transforms.FieldToJson$Value",
	"transforms.tagsToJson.originalField": "tags",
	"transforms.tagsToJson.destinationField": "tags",
	"transforms.injectSchemaKey.type": "com.redhat.insights.kafka.connect.transforms.InjectSchema$Key",
	"transforms.injectSchemaKey.schema": "{\"type\":\"string\",\"optional\":false, \"name\": \"com.redhat.cloud.inventory.syndication.pgtype=uuid\"}",
	"transforms.injectSchemaValue.type": "com.redhat.insights.kafka.connect.transforms.InjectSchema$Value",
	"transforms.injectSchemaValue.schema": "{\"type\":\"struct\",\"fields\":[{\"type\":\"string\",\"optional\":false,\"field\":\"account\"},{\"type\":\"string\",\"optional\":false,\"field\":\"display_name\"},{\"type\":\"string\",\"optional\":false,\"field\":\"tags\", \"name\": \"com.redhat.cloud.inventory.syndication.pgtype=jsonb\"},{\"type\":\"string\",\"optional\":false,\"field\":\"updated\", \"name\": \"com.redhat.cloud.inventory.syndication.pgtype=timestamptz\"},{\"type\":\"string\",\"optional\":false,\"field\":\"created\", \"name\": \"com.redhat.cloud.inventory.syndication.pgtype=timestamptz\"},{\"type\":\"string\",\"optional\":false,\"field\":\"stale_timestamp\", \"name\": \"com.redhat.cloud.inventory.syndication.pgtype=timestamptz\"},{\"type\":\"string\",\"optional\":false,\"field\":\"system_profile\", \"name\": \"com.redhat.cloud.inventory.syndication.pgtype=jsonb\"},{\"type\":\"string\",\"optional\":true,\"field\":\"insights_id\"}],\"optional\":false}",

	"errors.tolerance": "all",
	"errors.deadletterqueue.topic.name": "platform.cyndi.dlq",
	"errors.deadletterqueue.topic.replication.factor": 1,
	"errors.deadletterqueue.context.headers.enable":true,
	"errors.retry.delay.max.ms": 60000,
	"errors.retry.timeout": 600000,
	"errors.log.enable":true,
	"errors.log.include.messages":true,

	"connection.attempts": 60,
	"connection.backoff.ms": 10000
}`

const defaultConnectorTasksMax int64 = 16
const defaultConnectorBatchSize int64 = 100
const defaultConnectorMaxAge int64 = 45
const defaultAllowlistSystemProfile = "sap_system,sap_sids"

const defaultDBTableInitScript = `
CREATE TABLE inventory.{{.TableName}} (
	id uuid PRIMARY KEY,
	account character varying(10) NOT NULL,
	display_name character varying(200) NOT NULL,
	tags jsonb NOT NULL,
	updated timestamp with time zone NOT NULL,
	created timestamp with time zone NOT NULL,
	stale_timestamp timestamp with time zone NOT NULL,
	system_profile jsonb NOT NULL,
	insights_id uuid
);

CREATE INDEX {{.TableName}}_account_index ON inventory.{{.TableName}}
(account);

CREATE INDEX {{.TableName}}_display_name_index ON inventory.{{.TableName}}
(display_name);

CREATE INDEX {{.TableName}}_tags_index ON inventory.{{.TableName}} USING GIN
(tags JSONB_PATH_OPS);

CREATE INDEX {{.TableName}}_stale_timestamp_index ON
inventory.{{.TableName}} (stale_timestamp);

CREATE INDEX {{.TableName}}_system_profile_index ON inventory.{{.TableName}}
USING GIN (system_profile JSONB_PATH_OPS);

CREATE INDEX {{.TableName}}_insights_id_index ON
inventory.{{.TableName}} (insights_id);
`

const defaultStandardInterval int64 = 120

var defaultValidationConfig = ValidationConfiguration{
	Interval:            60 * 30,
	AttemptsThreshold:   3,
	PercentageThreshold: 5,
}

var defaultValidationConfigInit = ValidationConfiguration{
	Interval:            60,
	AttemptsThreshold:   30,
	PercentageThreshold: 5,
}
