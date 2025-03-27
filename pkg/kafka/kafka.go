package kafka

type Config struct {
	BrokerID  string `yaml:"BROKER" env:"KAFKA_BROKER_ID" env-default:"1"`
	Zookeeper string `yaml:"ZOOKEEPER" env:"KAFKA_ZOOKEEPER_CONNECT" env-default:"zookeeper:2181"`
	Listener  string `yaml:"LISTENER" env:"KAFKA_LISTENER_NAME" env-default:"PLAIN"`
	Port      string `yaml:"PORT" env:"KAFKA_LISTENER_PLAIN_PORT" env-default:"9092"`
	Address   string `yaml:"ADDRESS" env:"KAFKA_ADVERTISED_LISTENERS" env-default:"PLAIN://localhost:9092"`
	Replicas  string `yaml:"REPLICAS" env:"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR" env-default:"1"`
}
