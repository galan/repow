package config

type config struct {
	Options options `koanf:"options"`
	Server  server  `koanf:"server"`
	Gitlab  gitlab  `koanf:"gitlab"`
	Slack   slack   `koanf:"slack"`
}

type options struct {
	Style            string `koanf:"style"`
	Parallelism      int    `koanf:"parallelism"`
	OptionalContacts bool   `koanf:"optionalcontacts"`
}

type server struct {
	Port int `koanf:"port"`
}

type gitlab struct {
	Host               string `koanf:"host"`
	ApiToken           string `koanf:"apitoken"`
	DownloadRetryCount int    `koanf:"downloadretrycount"`
	SecretToken        string `koanf:"secrettoken"`
}
type slack struct {
	Token     string `koanf:"token"`
	ChannelId string `koanf:"channelid"`
	Prefix    string `koanf:"prefix"`
}
