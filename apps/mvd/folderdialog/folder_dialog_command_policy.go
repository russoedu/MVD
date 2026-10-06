package folderdialog

import (
	"strings"

	"youtube-downloader/apps/mvd/oscommand"
)

// startEnv is the variable the Windows script reads the starting folder from.
const startEnv = "MVD_START"

// windowsScript shows the Explorer style folder chooser (the one with an address bar
// and the folder tree, as in File Explorer) above every other window and prints the
// chosen path as UTF-8. Printing nothing means the person cancelled. If that chooser
// cannot be created it falls back to the classic tree-only one.
const windowsScript = `$ErrorActionPreference = 'Stop'
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
Add-Type -AssemblyName System.Windows.Forms
$owner = New-Object System.Windows.Forms.Form
$owner.TopMost = $true
$owner.ShowInTaskbar = $false
$owner.WindowState = 'Minimized'
$owner.Show()
$start = $env:MVD_START
if (-not ($start -and (Test-Path -LiteralPath $start -PathType Container))) { $start = '' }
$path = $null
try {
  Add-Type -TypeDefinition @'
using System;
using System.Runtime.InteropServices;

public static class MvdFolderPicker {
  [ComImport, Guid("DC1C5A9C-E88A-4dde-A5A1-60F82A20AEF7")]
  private class FileOpenDialogClass { }

  [ComImport, Guid("42F85136-DB7E-439C-85F1-E4075D135FC8"), InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
  private interface IFileDialog {
    [PreserveSig] int Show(IntPtr parent);
    void SetFileTypes(uint count, IntPtr specs);
    void SetFileTypeIndex(uint index);
    void GetFileTypeIndex(out uint index);
    void Advise(IntPtr events, out uint cookie);
    void Unadvise(uint cookie);
    void SetOptions(uint options);
    void GetOptions(out uint options);
    void SetDefaultFolder(IShellItem item);
    void SetFolder(IShellItem item);
    void GetFolder(out IShellItem item);
    void GetCurrentSelection(out IShellItem item);
    void SetFileName([MarshalAs(UnmanagedType.LPWStr)] string name);
    void GetFileName([MarshalAs(UnmanagedType.LPWStr)] out string name);
    void SetTitle([MarshalAs(UnmanagedType.LPWStr)] string title);
    void SetOkButtonLabel([MarshalAs(UnmanagedType.LPWStr)] string text);
    void SetFileNameLabel([MarshalAs(UnmanagedType.LPWStr)] string label);
    void GetResult(out IShellItem item);
    void AddPlace(IShellItem item, int placement);
    void SetDefaultExtension([MarshalAs(UnmanagedType.LPWStr)] string extension);
    void Close(int result);
    void SetClientGuid(ref Guid guid);
    void ClearClientData();
    void SetFilter(IntPtr filter);
  }

  [ComImport, Guid("43826D1E-E718-42EE-BC55-A1E261C37BFE"), InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
  private interface IShellItem {
    void BindToHandler(IntPtr context, ref Guid handler, ref Guid iid, out IntPtr result);
    void GetParent(out IShellItem item);
    void GetDisplayName(uint form, [MarshalAs(UnmanagedType.LPWStr)] out string name);
    void GetAttributes(uint mask, out uint attributes);
    void Compare(IShellItem item, uint hint, out int order);
  }

  [DllImport("shell32.dll", CharSet = CharSet.Unicode, PreserveSig = false)]
  private static extern void SHCreateItemFromParsingName(string path, IntPtr context, ref Guid iid, out IShellItem item);

  // Shows the Explorer style folder chooser. Returns the folder, or null if it was cancelled.
  public static string Pick(IntPtr owner, string start) {
    IFileDialog dialog = (IFileDialog)new FileOpenDialogClass();
    uint options;
    dialog.GetOptions(out options);
    dialog.SetOptions(options | 0x20 | 0x40 | 0x800); // pick folders, real folders only, must exist
    dialog.SetTitle("Choose a folder");
    if (!string.IsNullOrEmpty(start)) {
      Guid iid = new Guid("43826D1E-E718-42EE-BC55-A1E261C37BFE");
      IShellItem folder;
      SHCreateItemFromParsingName(start, IntPtr.Zero, ref iid, out folder);
      dialog.SetFolder(folder);
    }
    int hr = dialog.Show(owner);
    if (hr == unchecked((int)0x800704C7)) return null; // cancelled
    if (hr != 0) Marshal.ThrowExceptionForHR(hr);
    IShellItem item;
    dialog.GetResult(out item);
    string path;
    item.GetDisplayName(0x80058000, out path); // SIGDN_FILESYSPATH
    return path;
  }
}
'@
  $path = [MvdFolderPicker]::Pick($owner.Handle, $start)
} catch {
  # No Explorer style chooser here: fall back to the classic one.
  $dialog = New-Object System.Windows.Forms.FolderBrowserDialog
  $dialog.Description = 'Choose a folder'
  $dialog.ShowNewFolderButton = $true
  if ($start) { $dialog.SelectedPath = $start }
  if ($dialog.ShowDialog($owner) -eq [System.Windows.Forms.DialogResult]::OK) { $path = $dialog.SelectedPath }
}
$owner.Close()
if ($path) { [Console]::Out.Write($path) }
`

// folderDialogCommand picks the way to show a folder chooser on goos, using only the
// programs has reports as installed. ok is false when there is none.
func folderDialogCommand(goos, start string, has func(string) bool) (cmd oscommand.Command, ok bool) {
	switch goos {
	case "windows":
		if !has("powershell") {
			return oscommand.Command{}, false
		}

		return oscommand.Command{
			Name: "powershell",
			Args: []string{"-NoProfile", "-NonInteractive", "-STA", "-EncodedCommand", oscommand.EncodePowerShell(windowsScript)},
			Env:  []string{startEnv + "=" + start},
		}, true
	case "darwin":
		if !has("osascript") {
			return oscommand.Command{}, false
		}
		// The folder is argv, not part of the AppleScript text.
		script := []string{
			"-e", "on run argv",
			"-e", `if (count of argv) > 0 then`,
			"-e", `return POSIX path of (choose folder with prompt "Choose a folder" default location (POSIX file (item 1 of argv)))`,
			"-e", "end if",
			"-e", `return POSIX path of (choose folder with prompt "Choose a folder")`,
			"-e", "end run",
		}
		if start != "" {
			script = append(script, start)
		}

		return oscommand.Command{Name: "osascript", Args: script}, true
	default:
		switch {
		case has("zenity"):
			args := []string{"--file-selection", "--directory", "--title=Choose a folder"}
			if start != "" {
				args = append(args, "--filename="+strings.TrimRight(start, "/")+"/")
			}

			return oscommand.Command{Name: "zenity", Args: args}, true
		case has("kdialog"):
			args := []string{"--getexistingdirectory"}
			if start != "" {
				args = append(args, start)
			}

			return oscommand.Command{Name: "kdialog", Args: args}, true
		}

		return oscommand.Command{}, false
	}
}
