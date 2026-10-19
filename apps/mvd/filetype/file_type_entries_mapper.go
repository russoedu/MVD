package filetype

import "strings"

// RegistryValue is one value to write in the Windows registry under HKEY_CURRENT_USER. An
// empty Name is the default value of the key.
type RegistryValue struct {
	Key, Name, Value string
}

// registryClasses is where Windows keeps what the person has chosen for each kind of file.
const registryClasses = `Software\Classes\`

// RegistryValues are what Windows needs to know for the program at path: the extension is
// a download list, shown with the icon of the program, and opened by running the program
// with the file as its one argument.
func RegistryValues(program string) []RegistryValue {
	kind := registryClasses + ProgramID
	return []RegistryValue{
		{Key: registryClasses + Extension, Value: ProgramID},
		{Key: kind, Value: Description},
		{Key: kind + `\DefaultIcon`, Value: `"` + program + `",0`},
		{Key: kind + `\shell\open\command`, Value: `"` + program + `" "%1"`},
	}
}

// RegistryKeys are the keys RegistryValues writes, which are the ones taken away again.
func RegistryKeys() []string {
	return []string{registryClasses + Extension, registryClasses + ProgramID}
}

// MimeXML is the description of the kind of file for Linux's shared mime database: its
// name, the extension that tells it, and the icon of the app.
func MimeXML() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<mime-info xmlns="http://www.freedesktop.org/standards/shared-mime-info">
  <mime-type type="` + MimeType + `">
    <comment>` + Description + `</comment>
    <glob pattern="*` + Extension + `"/>
    <icon name="mvd"/>
  </mime-type>
</mime-info>
`
}

// ListInArgs finds the download list a start was given: the first argument that ends in
// the extension, whatever its case. "" when there is none.
func ListInArgs(args []string) string {
	for _, a := range args {
		if strings.HasSuffix(strings.ToLower(a), Extension) {
			return a
		}
	}
	return ""
}
