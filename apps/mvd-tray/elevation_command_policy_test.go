package main

import (
	"regexp"
	"strings"
	"testing"
)

func TestAQuoteInAPathCannotEndTheQuotedString(t *testing.T) {
	cases := map[string]string{
		`C:\plain\path`:                      `'C:\plain\path'`,
		`C:\it's here\mvd-tray.exe`:          `'C:\it''s here\mvd-tray.exe'`,
		`C:\x'; Remove-Item -Recurse C:\ ;'`: `'C:\x''; Remove-Item -Recurse C:\ ;'''`,
	}
	for in, want := range cases {
		if got := quotePowerShell(in); got != want {
			t.Errorf("quotePowerShell(%q) = %s, want %s", in, got, want)
		}
	}
}

func TestTheElevatedScriptInstallsOnlyTheThreePathsItIsGiven(t *testing.T) {
	script := elevatedInstallScript(`C:\Downloads\mvd-tray (1).exe`, `C:\Program Files\MVD\mvd-tray.exe`, `C:\ProgramData\Menu\MVD.lnk`)

	for _, want := range []string{
		`$source = 'C:\Downloads\mvd-tray (1).exe'`,
		`$dest = 'C:\Program Files\MVD\mvd-tray.exe'`,
		`$link = 'C:\ProgramData\Menu\MVD.lnk'`,
		"Copy-Item -LiteralPath $source -Destination $dest -Force",
		"$shortcut.TargetPath = $dest",
		"$ErrorActionPreference = 'Stop'",
	} {
		if !strings.Contains(script, want) {
			t.Errorf("the script lacks %q:\n%s", want, script)
		}
	}
}

func TestTheElevationCommandAsksForAdministratorApprovalThroughPowerShellAndCarriesTheScript(t *testing.T) {
	source, dest, link := `C:\Downloads\mvd.exe`, `C:\Program Files\MVD\mvd-tray.exe`, `C:\ProgramData\MVD.lnk`

	cmd := elevationCommand(source, dest, link)

	if cmd.name != "powershell" {
		t.Fatalf("name = %q", cmd.name)
	}
	outer := decodePowerShell(t, cmd.args)
	for _, want := range []string{"-Verb RunAs", "-Wait", "-PassThru", "exit $step.ExitCode", "exit 1223"} {
		if !strings.Contains(outer, want) {
			t.Errorf("the outer script lacks %q:\n%s", want, outer)
		}
	}

	found := regexp.MustCompile(`'-EncodedCommand', '([A-Za-z0-9+/=]+)'`).FindStringSubmatch(outer)
	if found == nil {
		t.Fatalf("the outer script does not carry an encoded script:\n%s", outer)
	}
	inner := decodePowerShell(t, []string{"-EncodedCommand", found[1]})
	if inner != elevatedInstallScript(source, dest, link) {
		t.Errorf("the elevated script is not the install script for those paths:\n%s", inner)
	}
}

func TestNothingTheCallerSuppliesCanReachTheOuterScriptExceptAsEncodedData(t *testing.T) {
	cmd := elevationCommand(`C:\x'; Remove-Item -Recurse C:\ #`, `C:\d`, `C:\l`)

	if got := strings.Join(cmd.args, " "); strings.Contains(got, "Remove-Item") {
		t.Errorf("a path reached the command line: %s", got)
	}
	if outer := decodePowerShell(t, cmd.args); strings.Contains(outer, "Remove-Item") {
		t.Errorf("a path reached the outer script as text:\n%s", outer)
	}
}
