#define MyAppName "EB Gestión"
#define MyAppVersion "0.9.0 Beta Operativa OP1"
#define MyAppPublisher "Restaurante Entre Bahías"
#define MyAppExeName "EBGestion.exe"

[Setup]
AppId={{98A4D582-7566-4E57-B35F-81D4706CFC2B}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher={#MyAppPublisher}
DefaultDirName={localappdata}\Programs\EB Gestion
DefaultGroupName=EB Gestión
DisableProgramGroupPage=yes
PrivilegesRequired=lowest
OutputDir=..\dist
OutputBaseFilename=EBGestion-Setup
Compression=lzma2
SolidCompression=yes
WizardStyle=modern
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
UninstallDisplayIcon={app}\{#MyAppExeName}

CloseApplications=force
CloseApplicationsFilter=*.exe
RestartApplications=no
VersionInfoVersion=0.9.0.1
VersionInfoDescription=EB Gestión Beta Operativa OP1

[Files]
Source: "..\dist\EBGestion.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\dist\LEEME-DESKTOP.txt"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{autoprograms}\EB Gestión"; Filename: "{app}\{#MyAppExeName}"
Name: "{autodesktop}\EB Gestión"; Filename: "{app}\{#MyAppExeName}"; Tasks: desktopicon

[Tasks]
Name: "desktopicon"; Description: "Crear acceso directo en el escritorio"; GroupDescription: "Accesos directos:"; Flags: checkedonce

[Run]
Filename: "{app}\{#MyAppExeName}"; Description: "Abrir EB Gestión"; Flags: nowait postinstall skipifsilent
