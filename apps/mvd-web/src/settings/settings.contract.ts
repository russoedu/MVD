// The settings as the server sends them (libs/mvd-server/settings).

export interface Settings {
  outputDir:                  string
  videoQuality:               string
  audioQuality:               string
  mergeOutputFormat:          string
  outputTemplate:             string
  maxConcurrentDownloads:     number
  /** 0 turns parallel fragments off. */
  concurrentFragments:        number
  downloadOfficialMusicVideo: boolean
  autoRetry:                  boolean
  /** "all", "off", or one browser such as "firefox". */
  cookies:                    string
  createLogFile:              boolean
  logDir:                     string
}

export interface SettingsOptions {
  videoQualities: string[]
  audioQualities: string[]
  mergeFormats:   string[]
  /** Browsers found on this machine. */
  browsers:       string[]
}

export interface SettingsDocument {
  settings: Settings
  options:  SettingsOptions
}

export interface PickedFolder {
  path:      string
  cancelled: boolean
}
