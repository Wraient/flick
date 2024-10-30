[Setup]
AppName=Flick Installer
AppVersion=0.0.1
DefaultDirName={pf}\Flick
DefaultGroupName=Flick
AllowNoIcons=yes
OutputBaseFilename=FlickInstaller
UsePreviousAppDir=yes
Compression=lzma2
SolidCompression=yes

[Tasks]
; Define a task for creating a desktop shortcut
Name: "desktopicon"; Description: "Create a &desktop shortcut"; GroupDescription: "Additional Options";

[Files]
; Copy the Flick executable to the install directory
Source: "Z:releases/flick-0.0.1/windows/flick.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "mpv.exe"; DestDir: "{app}\bin"; Flags: ignoreversion

[Icons]
; Create the application icon in the Start Menu
Name: "{group}\Flick"; Filename: "{app}\flick.exe"
; Create a desktop shortcut if the user checked the option
Name: "{userdesktop}\Flick"; Filename: "{app}\flick.exe"; Tasks: desktopicon
