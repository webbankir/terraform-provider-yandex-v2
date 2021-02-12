package yandex

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/kafka/v1"
	"google.golang.org/genproto/protobuf/field_mask"
)

const (
	yandexMDBKafkaClusterCreateTimeout = 60 * time.Minute
	yandexMDBKafkaClusterReadTimeout   = 5 * time.Minute
	yandexMDBKafkaClusterDeleteTimeout = 60 * time.Minute
	yandexMDBKafkaClusterUpdateTimeout = 60 * time.Minute
)

func resourceYandexMDBKafkaCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexMDBKafkaClusterCreate,
		Read:   resourceYandexMDBKafkaClusterRead,
		Update: resourceYandexMDBKafkaClusterUpdate,
		Delete: resourceYandexMDBKafkaClusterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexMDBKafkaClusterCreateTimeout),
			Read:   schema.DefaultTimeout(yandexMDBKafkaClusterReadTimeout),
			Update: schema.DefaultTimeout(yandexMDBKafkaClusterUpdateTimeout),
			Delete: schema.DefaultTimeout(yandexMDBKafkaClusterDeleteTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"network_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem:     resourceYandexMDBKafkaClusterConfig(),
			},
			"environment": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "PRODUCTION",
				ValidateFunc: validateParsableValue(parseKafkaEnv),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"subnet_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"topic": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     resourceYandexMDBKafkaTopic(),
			},
			"user": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      kafkaUserHash,
				Elem:     resourceYandexMDBKafkaUser(),
			},
			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"security_group_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Optional: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"health": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceYandexMDBKafkaClusterConfig() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"version": {
				Type:     schema.TypeString,
				Required: true,
			},
			"zones": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
			},
			"kafka": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem:     resourceYandexMDBKafkaClusterKafkaConfig(),
			},
			"brokers_count": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},
			"assign_public_ip": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"zookeeper": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem:     resourceYandexMDBKafkaClusterZookeeperConfig(),
			},
		},
	}
}

func resourceYandexMDBKafkaClusterResources() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"resource_preset_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"disk_size": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"disk_type_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceYandexMDBKafkaTopic() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"partitions": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"replication_factor": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"topic_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem:     resourceYandexMDBKafkaClusterTopicConfig(),
			},
		},
	}
}

func resourceYandexMDBKafkaUser() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"permission": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      kafkaUserPermissionHash,
				Elem:     resourceYandexMDBKafkaPermission(),
			},
		},
	}
}

func resourceYandexMDBKafkaPermission() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"topic_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"role": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceYandexMDBKafkaClusterKafkaConfig() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"resources": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem:     resourceYandexMDBKafkaClusterResources(),
			},
			"kafka_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem:     resourceYandexMDBKafkaClusterKafkaSettings(),
			},
		},
	}
}

func resourceYandexMDBKafkaClusterKafkaSettings() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"compression_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateParsableValue(parseKafkaCompression),
			},
			"log_flush_interval_messages": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"log_flush_interval_ms": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"log_flush_scheduler_interval_ms": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"log_retention_bytes": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"log_retention_hours": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"log_retention_minutes": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"log_retention_ms": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"log_segment_bytes": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"log_preallocate": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func resourceYandexMDBKafkaClusterTopicConfig() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"cleanup_policy": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateParsableValue(parseKafkaTopicCleanupPolicy),
			},
			"compression_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateParsableValue(parseKafkaCompression),
			},
			"delete_retention_ms": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"file_delete_delay_ms": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"flush_messages": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"flush_ms": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"min_compaction_lag_ms": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"retention_bytes": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"retention_ms": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"max_message_bytes": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"min_insync_replicas": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"segment_bytes": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"preallocate": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func resourceYandexMDBKafkaClusterZookeeperConfig() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"resources": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem:     resourceYandexMDBKafkaClusterResources(),
			},
		},
	}
}

func resourceYandexMDBKafkaClusterCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	req, err := prepareKafkaCreateRequest(d, config)

	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	log.Printf("[DEBUG] Creating Kafka cluster: %+v", req)

	op, err := config.sdk.WrapOperation(config.sdk.MDB().Kafka().Cluster().Create(ctx, req))
	if err != nil {
		return fmt.Errorf("error while requesting API to create Kafka Cluster: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("error while getting Kafka create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*kafka.CreateClusterMetadata)
	if !ok {
		return fmt.Errorf("could not get Cluster ID from create operation metadata")
	}

	d.SetId(md.ClusterId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while waiting for operation to create Kafka Cluster: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("kafka cluster creation failed: %s", err)
	}
	log.Printf("[DEBUG] Finished creating Kafka cluster %q", md.ClusterId)

	return resourceYandexMDBKafkaClusterRead(d, meta)
}

// Returns request for creating the Cluster.
func prepareKafkaCreateRequest(d *schema.ResourceData, meta *Config) (*kafka.CreateClusterRequest, error) {
	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return nil, fmt.Errorf("error while expanding labels on Kafka Cluster create: %s", err)
	}

	folderID, err := getFolderID(d, meta)
	if err != nil {
		return nil, fmt.Errorf("error getting folder ID while creating Kafka Cluster: %s", err)
	}

	e := d.Get("environment").(string)
	env, err := parseKafkaEnv(e)
	if err != nil {
		return nil, fmt.Errorf("error resolving environment while creating Kafka Cluster: %s", err)
	}

	configSpec, err := expandKafkaConfigSpec(d)
	if err != nil {
		return nil, fmt.Errorf("error while expanding configuration on Kafka Cluster create: %s", err)
	}

	subnets := []string{}
	if v, ok := d.GetOk("subnet_ids"); ok {
		for _, subnet := range v.([]interface{}) {
			subnets = append(subnets, subnet.(string))
		}
	}

	topicSpecs, err := expandKafkaTopics(d)
	if err != nil {
		return nil, fmt.Errorf("error while expanding topics on Kafka Cluster create: %s", err)
	}

	userSpecs, err := expandKafkaUsers(d)
	if err != nil {
		return nil, fmt.Errorf("error while expanding users on Kafka Cluster create: %s", err)
	}

	securityGroupIds := expandSecurityGroupIds(d.Get("security_group_ids"))

	req := kafka.CreateClusterRequest{
		FolderId:         folderID,
		Name:             d.Get("name").(string),
		Description:      d.Get("description").(string),
		NetworkId:        d.Get("network_id").(string),
		Environment:      env,
		ConfigSpec:       configSpec,
		Labels:           labels,
		SubnetId:         subnets,
		TopicSpecs:       topicSpecs,
		UserSpecs:        userSpecs,
		SecurityGroupIds: securityGroupIds,
	}
	return &req, nil
}

func resourceYandexMDBKafkaClusterRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutRead))
	defer cancel()

	cluster, err := config.sdk.MDB().Kafka().Cluster().Get(ctx, &kafka.GetClusterRequest{
		ClusterId: d.Id(),
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Cluster %q", d.Get("name").(string)))
	}

	createdAt, err := getTimestamp(cluster.CreatedAt)
	if err != nil {
		return err
	}

	d.Set("created_at", createdAt)
	d.Set("name", cluster.Name)
	d.Set("folder_id", cluster.FolderId)
	d.Set("network_id", cluster.NetworkId)
	d.Set("environment", cluster.GetEnvironment().String())
	d.Set("health", cluster.GetHealth().String())
	d.Set("status", cluster.GetStatus().String())
	d.Set("description", cluster.Description)

	cfg, err := flattenKafkaConfig(cluster)
	if err != nil {
		return err
	}
	if err := d.Set("config", cfg); err != nil {
		return err
	}

	topics, err := listKafkaTopics(ctx, config, d.Id())
	if err != nil {
		return err
	}

	topicSpecs, err := expandKafkaTopics(d)
	if err != nil {
		return err
	}
	sortKafkaTopics(topics, topicSpecs)

	if err := d.Set("topic", flattenKafkaTopics(topics)); err != nil {
		return err
	}

	dUsers, err := expandKafkaUsers(d)
	if err != nil {
		return err
	}
	passwords := kafkaUsersPasswords(dUsers)

	users, err := listKafkaUsers(ctx, config, d.Id())
	if err != nil {
		return err
	}
	if err := d.Set("user", flattenKafkaUsers(users, passwords)); err != nil {
		return err
	}

	if err := d.Set("security_group_ids", cluster.SecurityGroupIds); err != nil {
		return err
	}

	return d.Set("labels", cluster.Labels)
}

func resourceYandexMDBKafkaClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Updating Kafka Cluster %q", d.Id())

	d.Partial(true)

	if err := updateKafkaClusterParams(d, meta); err != nil {
		return err
	}

	if d.HasChange("topic") {
		if err := updateKafkaClusterTopics(d, meta); err != nil {
			return err
		}
	}

	if d.HasChange("user") {
		if err := updateKafkaClusterUsers(d, meta); err != nil {
			return err
		}
	}

	d.Partial(false)

	log.Printf("[DEBUG] Finished updating Kafka Cluster %q", d.Id())
	return resourceYandexMDBKafkaClusterRead(d, meta)
}

func resourceYandexMDBKafkaClusterDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting Kafka Cluster %q", d.Id())

	req := &kafka.DeleteClusterRequest{
		ClusterId: d.Id(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.MDB().Kafka().Cluster().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Kafka Cluster %q", d.Get("name").(string)))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting Kafka Cluster %q", d.Id())
	return nil
}

func listKafkaTopics(ctx context.Context, config *Config, id string) ([]*kafka.Topic, error) {
	ret := []*kafka.Topic{}
	pageToken := ""
	for {
		resp, err := config.sdk.MDB().Kafka().Topic().List(ctx, &kafka.ListTopicsRequest{
			ClusterId: id,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("error while getting list of topics for '%s': %s", id, err)
		}
		ret = append(ret, resp.Topics...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return ret, nil
}

func listKafkaUsers(ctx context.Context, config *Config, id string) ([]*kafka.User, error) {
	ret := []*kafka.User{}
	pageToken := ""
	for {
		resp, err := config.sdk.MDB().Kafka().User().List(ctx, &kafka.ListUsersRequest{
			ClusterId: id,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("error while getting list of users for '%s': %s", id, err)
		}
		ret = append(ret, resp.Users...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return ret, nil
}

var mdbKafkaUpdateFieldsMap = map[string]string{
	"name":                   "name",
	"description":            "description",
	"labels":                 "labels",
	"security_group_ids":     "security_group_ids",
	"config.0.zones":         "config_spec.zone_id",
	"config.0.brokers_count": "config_spec.brokers_count",
	"config.0.kafka.0.resources.0.resource_preset_id":                 "config_spec.kafka.resources.resource_preset_id",
	"config.0.kafka.0.resources.0.disk_type_id":                       "config_spec.kafka.resources.disk_type_id",
	"config.0.kafka.0.resources.0.disk_size":                          "config_spec.kafka.resources.disk_size",
	"config.0.kafka.0.kafka_config.0.compression_type":                "config_spec.kafka.kafka_config_{version}.compression_type",
	"config.0.kafka.0.kafka_config.0.log_flush_interval_messages":     "config_spec.kafka.kafka_config_{version}.log_flush_interval_messages",
	"config.0.kafka.0.kafka_config.0.log_flush_interval_ms":           "config_spec.kafka.kafka_config_{version}.log_flush_interval_ms",
	"config.0.kafka.0.kafka_config.0.log_flush_scheduler_interval_ms": "config_spec.kafka.kafka_config_{version}.log_flush_scheduler_interval_ms",
	"config.0.kafka.0.kafka_config.0.log_retention_bytes":             "config_spec.kafka.kafka_config_{version}.log_retention_bytes",
	"config.0.kafka.0.kafka_config.0.log_retention_hours":             "config_spec.kafka.kafka_config_{version}.log_retention_hours",
	"config.0.kafka.0.kafka_config.0.log_retention_minutes":           "config_spec.kafka.kafka_config_{version}.log_retention_minutes",
	"config.0.kafka.0.kafka_config.0.log_retention_ms":                "config_spec.kafka.kafka_config_{version}.log_retention_ms",
	"config.0.kafka.0.kafka_config.0.log_segment_bytes":               "config_spec.kafka.kafka_config_{version}.log_segment_bytes",
	"config.0.kafka.0.kafka_config.0.log_preallocate":                 "config_spec.kafka.kafka_config_{version}.log_preallocate",
	"config.0.zookeeper.0.resources.0.resource_preset_id":             "config_spec.zookeeper.resources.resource_preset_id",
	"config.0.zookeeper.0.resources.0.disk_type_id":                   "config_spec.zookeeper.resources.disk_type_id",
	"config.0.zookeeper.0.resources.0.disk_size":                      "config_spec.zookeeper.resources.disk_size",
}

func kafkaClusterUpdateRequest(d *schema.ResourceData) (*kafka.UpdateClusterRequest, error) {
	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return nil, fmt.Errorf("error expanding labels while updating Kafka cluster: %s", err)
	}

	configSpec, err := expandKafkaConfigSpec(d)
	if err != nil {
		return nil, fmt.Errorf("error expanding configSpec while updating Kafka cluster: %s", err)
	}

	req := &kafka.UpdateClusterRequest{
		ClusterId:        d.Id(),
		Name:             d.Get("name").(string),
		Description:      d.Get("description").(string),
		Labels:           labels,
		ConfigSpec:       configSpec,
		SecurityGroupIds: expandSecurityGroupIds(d.Get("security_group_ids")),
	}
	return req, nil
}

func getSuffixVerion(d *schema.ResourceData) string {
	return strings.Replace(d.Get("config.0.version").(string), ".", "_", -1)
}

func updateKafkaClusterParams(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	req, err := kafkaClusterUpdateRequest(d)
	if err != nil {
		return err
	}

	onDone := []func(){}
	updatePath := []string{}
	for field, path := range mdbKafkaUpdateFieldsMap {
		if d.HasChange(field) {
			updatePath = append(updatePath, strings.Replace(path, "{version}", getSuffixVerion(d), -1))
			onDone = append(onDone, func() {
				d.SetPartial(field)
			})
		}
	}

	if len(updatePath) == 0 {
		return nil
	}

	req.UpdateMask = &field_mask.FieldMask{Paths: updatePath}
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.MDB().Kafka().Cluster().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("error while requesting API to update Kafka Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating Kafka Cluster %q: %s", d.Id(), err)
	}

	for _, f := range onDone {
		f()
	}
	return nil
}

func updateKafkaClusterTopics(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	currTopics, err := listKafkaTopics(ctx, config, d.Id())
	if err != nil {
		return err
	}
	targetTopics, err := expandKafkaTopics(d)
	if err != nil {
		return err
	}
	sortKafkaTopics(currTopics, targetTopics)

	var toAdd []string
	toDelete, toAddSpecs := kafkaTopicsDiff(currTopics, targetTopics)

	for _, topic := range toDelete {
		err := deleteKafkaTopic(ctx, config, d, topic)
		if err != nil {
			return err
		}
	}

	for _, topic := range toAddSpecs {
		err := createKafkaTopic(ctx, config, d, topic)
		toAdd = append(toAdd, topic.Name)
		if err != nil {
			return err
		}
	}

	version, ok := d.GetOk("config.0.version")
	if !ok {
		return fmt.Errorf("you must specify version of Kafka")
	}

	oldSpecs, newSpecs := d.GetChange("topic")
	changedTopics, err := kafkaChangedTopics(d, oldSpecs.([]interface{}), newSpecs.([]interface{}), version.(string))
	if err != nil {
		return err
	}
	// Deleted and created topics also looks like changed topics, so we need to filter then manually
	// Remove them from changed topics slice
	modifiedTopics := kafkaFilterModifiedTopics(changedTopics, toDelete, toAdd)
	for _, t := range modifiedTopics {
		err := updateKafkaTopic(ctx, config, d, t.topic.Name, t, version.(string))
		if err != nil {
			return err
		}
	}

	d.SetPartial("topic")
	return nil
}

func deleteKafkaTopic(ctx context.Context, config *Config, d *schema.ResourceData, topicName string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Kafka().Topic().Delete(ctx, &kafka.DeleteTopicRequest{
			ClusterId: d.Id(),
			TopicName: topicName,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to delete topic from Kafka Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting topic from Kafka Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func createKafkaTopic(ctx context.Context, config *Config, d *schema.ResourceData, topicSpec *kafka.TopicSpec) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Kafka().Topic().Create(ctx, &kafka.CreateTopicRequest{
			ClusterId: d.Id(),
			TopicSpec: topicSpec,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to create topic in Kafka Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while adding topic to Kafka Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func updateKafkaClusterUsers(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	currUsers, err := listKafkaUsers(ctx, config, d.Id())
	if err != nil {
		return err
	}

	targetUsers, err := expandKafkaUsers(d)
	if err != nil {
		return err
	}
	toDelete, toAdd := kafkaUsersDiff(currUsers, targetUsers)

	for _, user := range toDelete {
		err := deleteKafkaUser(ctx, config, d, user)
		if err != nil {
			return err
		}
	}
	for _, user := range toAdd {
		err := createKafkaUser(ctx, config, d, user)
		if err != nil {
			return err
		}
	}

	oldSpecs, newSpecs := d.GetChange("user")
	err = updateKafkaUsers(ctx, config, d, oldSpecs.(*schema.Set), newSpecs.(*schema.Set))
	if err != nil {
		return err
	}

	d.SetPartial("user")
	return nil
}

func deleteKafkaUser(ctx context.Context, config *Config, d *schema.ResourceData, userName string) error {
	log.Printf("[DEBUG] Deleting Kafka user %q within cluster %q", userName, d.Id())
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Kafka().User().Delete(ctx, &kafka.DeleteUserRequest{
			ClusterId: d.Id(),
			UserName:  userName,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to delete user from Kafka Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting user from Kafka Cluster %q: %s", d.Id(), err)
	}
	log.Printf("[DEBUG] Finished deleting Kafka user %q", userName)
	return nil
}

func createKafkaUser(ctx context.Context, config *Config, d *schema.ResourceData, userSpec *kafka.UserSpec) error {
	req := &kafka.CreateUserRequest{
		ClusterId: d.Id(),
		UserSpec:  userSpec,
	}
	log.Printf("[DEBUG] Creating Kafka user %q: %+v", userSpec.Name, req)

	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Kafka().User().Create(ctx, req),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to create user in Kafka Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while adding user to Kafka Cluster %q: %s", d.Id(), err)
	}
	log.Printf("[DEBUG] Finished creating Kafka user %q", userSpec.Name)
	return nil
}

func updateKafkaUsers(ctx context.Context, config *Config, d *schema.ResourceData, oldSpecs *schema.Set, newSpecs *schema.Set) error {
	m := map[string]*kafka.UserSpec{}
	for _, spec := range oldSpecs.List() {
		user, err := expandKafkaUser(spec.(map[string]interface{}))
		if err != nil {
			return err
		}
		m[user.Name] = user
	}
	for _, spec := range newSpecs.List() {
		user, err := expandKafkaUser(spec.(map[string]interface{}))
		if err != nil {
			return err
		}
		if u, ok := m[user.Name]; ok {
			updatePaths := make([]string, 0, 2)

			if user.Password != u.Password {
				updatePaths = append(updatePaths, "password")
			}

			if fmt.Sprintf("%v", user.Permissions) != fmt.Sprintf("%v", u.Permissions) {
				updatePaths = append(updatePaths, "permissions")
			}

			if len(updatePaths) > 0 {
				req := &kafka.UpdateUserRequest{
					ClusterId:   d.Id(),
					UserName:    user.Name,
					Password:    user.Password,
					Permissions: user.Permissions,
					UpdateMask:  &field_mask.FieldMask{Paths: updatePaths},
				}
				err = updateKafkaUser(ctx, config, d, req)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func updateKafkaUser(ctx context.Context, config *Config, d *schema.ResourceData, req *kafka.UpdateUserRequest) error {
	log.Printf("[DEBUG] Updating Kafka user %q: %+v", req.UserName, req)
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Kafka().User().Update(ctx, req),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to update user in Kafka Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating user in Kafka Cluster %q: %s", d.Id(), err)
	}
	log.Printf("[DEBUG] Finished updating Kafka user %q", req.UserName)
	return nil
}

var mdbKafkaUpdateTopicFieldsMap = map[string]string{
	"topic.%d.name":                                 "topic_spec.name",
	"topic.%d.partitions":                           "topic_spec.partitions",
	"topic.%d.replication_factor":                   "topic_spec.replication_factor",
	"topic.%d.topic_config.0.cleanup_policy":        "topic_spec.topic_config_{version}.cleanup_policy",
	"topic.%d.topic_config.0.compression_type":      "topic_spec.topic_config_{version}.compression_type",
	"topic.%d.topic_config.0.delete_retention_ms":   "topic_spec.topic_config_{version}.delete_retention_ms",
	"topic.%d.topic_config.0.file_delete_delay_ms":  "topic_spec.topic_config_{version}.file_delete_delay_ms",
	"topic.%d.topic_config.0.flush_messages":        "topic_spec.topic_config_{version}.flush_messages",
	"topic.%d.topic_config.0.flush_ms":              "topic_spec.topic_config_{version}.flush_ms",
	"topic.%d.topic_config.0.min_compaction_lag_ms": "topic_spec.topic_config_{version}.min_compaction_lag_ms",
	"topic.%d.topic_config.0.retention_bytes":       "topic_spec.topic_config_{version}.retention_bytes",
	"topic.%d.topic_config.0.retention_ms":          "topic_spec.topic_config_{version}.retention_ms",
	"topic.%d.topic_config.0.max_message_bytes":     "topic_spec.topic_config_{version}.max_message_bytes",
	"topic.%d.topic_config.0.min_insync_replicas":   "topic_spec.topic_config_{version}.min_insync_replicas",
	"topic.%d.topic_config.0.segment_bytes":         "topic_spec.topic_config_{version}.segment_bytes",
	"topic.%d.topic_config.0.preallocate":           "topic_spec.topic_config_{version}.preallocate",
}

func updateKafkaTopic(ctx context.Context, config *Config, d *schema.ResourceData, topicName string, topicSpec IndexedTopicSpec, version string) error {
	request := &kafka.UpdateTopicRequest{
		ClusterId: d.Id(),
		TopicName: topicName,
		TopicSpec: topicSpec.topic,
	}

	onDone := []func(){}
	updatePath := []string{}

	for field, path := range mdbKafkaUpdateTopicFieldsMap {
		fd := fmt.Sprintf(field, topicSpec.index)
		if d.HasChange(fd) {
			updatePath = append(updatePath, strings.Replace(path, "{version}", getSuffixVerion(d), -1))
			onDone = append(onDone, func() {
				d.SetPartial(field)
			})
		}
	}

	if len(updatePath) == 0 {
		return nil
	}

	request.UpdateMask = &field_mask.FieldMask{Paths: updatePath}

	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Kafka().Topic().Update(ctx, request),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to update topic in Kafka Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating topic in Kafka Cluster %q: %s", d.Id(), err)
	}

	for _, f := range onDone {
		f()
	}
	return nil
}

func sortKafkaTopics(topics []*kafka.Topic, specs []*kafka.TopicSpec) {
	for i, spec := range specs {
		for j := i + 1; j < len(topics); j++ {
			if spec.Name == topics[j].Name {
				topics[i], topics[j] = topics[j], topics[i]
				break
			}
		}
	}
}
