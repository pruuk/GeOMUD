package configs

type FilePaths struct {
	WebDomain        ConfigString `yaml:"WebDomain"`
	WebCDNLocation   ConfigString `yaml:"WebCDNLocation"`
	DataFiles        ConfigString `yaml:"DataFiles"`
	PublicHtml       ConfigString `yaml:"PublicHtml"`
	AdminHtml        ConfigString `yaml:"AdminHtml"`
	HttpsCertFile    ConfigString `yaml:"HttpsCertFile"`
	HttpsKeyFile     ConfigString `yaml:"HttpsKeyFile"`
	SSHHostKeyFile   ConfigString `yaml:"SSHHostKeyFile"`
	HttpsEmail       ConfigSecret `yaml:"HttpsEmail"`
	HttpsCacheDir    ConfigString `yaml:"HttpsCacheDir"`
	CarefulSaveFiles ConfigBool   `yaml:"CarefulSaveFiles"`
}

func (f *FilePaths) Validate() {

	// Ignore WebDomain
	// Ignore WebCDNLocation
	// Ignore PublicHtml
	// Ignore AdminHtml
	// Ignore CarefulSaveFiles

	if f.DataFiles == `` {
		f.DataFiles = `_datafiles/world/empty` // default
	}

	if f.HttpsCacheDir == `` {
		f.HttpsCacheDir = `_datafiles/tls` // default
	}

}

func GetFilePathsConfig() FilePaths {
	configDataLock.RLock()
	defer configDataLock.RUnlock()

	if !configData.validated {
		configData.Validate()
	}
	return configData.FilePaths
}
