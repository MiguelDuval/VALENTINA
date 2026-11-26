Unicode true
!include "MUI2.nsh"
!include "LogicLib.nsh"
!include "x64.nsh"
!include "FileFunc.nsh"
Name "VALENTINA"
OutFile "${OUTPUT_FILE}"
InstallDir "$LOCALAPPDATA\Programs\VALENTINA"
RequestExecutionLevel user
ManifestDPIAware true
CRCCheck force
SetCompressor /SOLID lzma
ShowInstDetails show
ShowUninstDetails show
VIProductVersion "${PRODUCT_VERSION}.0"
VIAddVersionKey "ProductName" "VALENTINA"
VIAddVersionKey "FileDescription" "VALENTINA per-user setup"
VIAddVersionKey "FileVersion" "${PRODUCT_VERSION}"
VIAddVersionKey "LegalCopyright" "See LICENSES"
!define REGKEY "Software\Microsoft\Windows\CurrentVersion\Uninstall\VALENTINA"
!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_LICENSE "${PAYLOAD}\LICENSES\LICENSE.huginn.txt"
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES
!insertmacro MUI_LANGUAGE "Russian"
!insertmacro MUI_LANGUAGE "English"
Var SetupMutex

!macro Initialize
  SetShellVarContext current
  SetRegView 64
  ${IfNot} ${RunningX64}
    SetErrorLevel 1
    Quit
  ${EndIf}
  System::Call 'kernel32::CreateMutexW(p 0, i 0, w "Local\VALENTINA-Installer") p .r0 ?e'
  Pop $1
  StrCpy $SetupMutex $0
  ${If} $1 = 183
    SetErrorLevel 1
    Quit
  ${EndIf}
!macroend

Function .onInit
  !insertmacro Initialize
FunctionEnd

; Refuse junctions/symlinks in the target ancestry and license directory.
; Otherwise an explicit file deletion could still traverse into user data.
!macro RejectFileLink PATH
  System::Call 'kernel32::GetFileAttributesW(w "${PATH}") i .r4'
  ${If} $4 != -1
    IntOp $4 $4 & 0x400
    ${If} $4 != 0
      SetErrorLevel 3
      Quit
    ${EndIf}
  ${EndIf}
!macroend
!macro CheckPaths
  !insertmacro RejectFileLink "$INSTDIR\VALENTINA.exe"
  !insertmacro RejectFileLink "$INSTDIR\Uninstall.exe"
  !insertmacro RejectFileLink "$INSTDIR\README.txt"
  !insertmacro RejectFileLink "$INSTDIR\VALENTINA-BUILD-MANIFEST.json"
  !insertmacro RejectFileLink "$INSTDIR\SHA256SUMS.txt"
  !insertmacro RejectFileLink "$INSTDIR\LICENSES\LICENSE.huginn.txt"
  !insertmacro RejectFileLink "$INSTDIR\LICENSES\THIRD-PARTY-NOTICES.txt"
  StrCpy $3 "$INSTDIR\LICENSES"
  path_loop:
    System::Call 'kernel32::GetFileAttributesW(w r3) i .r4'
    ${If} $4 != -1
      IntOp $4 $4 & 0x400
      ${If} $4 != 0
        SetErrorLevel 3
        Quit
      ${EndIf}
    ${EndIf}
    ${GetParent} "$3" $4
    ${If} $4 == ""
    ${OrIf} $4 == $3
      Goto paths_safe
    ${EndIf}
    StrCpy $3 $4
    Goto path_loop
  paths_safe:
!macroend

; Test write access without truncating the executable. Windows refuses write
; access to a mapped/running image. Never kill it or defer replacement to reboot.
!macro CheckUnlocked
  IfFileExists "$INSTDIR\VALENTINA.exe" 0 unlocked
  System::Call 'kernel32::CreateFileW(w "$INSTDIR\VALENTINA.exe", i 0x40000000, i 1, p 0, i 3, i 0, p 0) p .r0'
  ${If} $0 == -1
    MessageBox MB_OK|MB_ICONSTOP "Закройте VALENTINA. Файл приложения занят или недоступен для записи." /SD IDOK
    SetErrorLevel 2
    Quit
  ${EndIf}
  System::Call 'kernel32::CloseHandle(p r0)'
  unlocked:
!macroend

Section "VALENTINA"
  !insertmacro CheckPaths
  ; A second location must not take over another installation's registration.
  ReadRegStr $0 HKCU "${REGKEY}" "InstallLocation"
  ${If} $0 != ""
  ${AndIf} $0 != $INSTDIR
    MessageBox MB_OK|MB_ICONSTOP "VALENTINA уже установлена в другом каталоге. Используйте прежний каталог." /SD IDOK
    SetErrorLevel 3
    Quit
  ${EndIf}
  ${If} $0 == ""
    ; A fresh target must be empty; do not overwrite arbitrary existing files.
    FindFirst $1 $2 "$INSTDIR\*"
    check_entry:
      ${If} $2 == ""
        Goto empty_target
      ${EndIf}
      ${If} $2 != "."
      ${AndIf} $2 != ".."
        FindClose $1
        SetErrorLevel 3
        MessageBox MB_OK|MB_ICONSTOP "Выберите пустой каталог для установки VALENTINA." /SD IDOK
        Quit
      ${EndIf}
      FindNext $1 $2
      Goto check_entry
    empty_target:
    FindClose $1
    IfFileExists "$SMPROGRAMS\VALENTINA.lnk" 0 shortcut_free
      SetErrorLevel 3
      Quit
    shortcut_free:
  ${Else}
    IfFileExists "$INSTDIR\Uninstall.exe" +3 0
      SetErrorLevel 3
      Quit
  ${EndIf}
  !insertmacro CheckUnlocked
  SetOverwrite on
  ClearErrors
  SetOutPath "$INSTDIR"
  File "${PAYLOAD}\VALENTINA.exe"
  IfErrors install_failed
  File "${PAYLOAD}\README.txt"
  File "${PAYLOAD}\VALENTINA-BUILD-MANIFEST.json"
  File "${PAYLOAD}\SHA256SUMS.txt"
  SetOutPath "$INSTDIR\LICENSES"
  File "${PAYLOAD}\LICENSES\LICENSE.huginn.txt"
  File "${PAYLOAD}\LICENSES\THIRD-PARTY-NOTICES.txt"
  SetOutPath "$INSTDIR"
  WriteUninstaller "$INSTDIR\Uninstall.exe"
  IfErrors install_failed
  ; Use a data-independent working directory; no mutable state in Programs.
  SetOutPath "$PROFILE"
  CreateShortcut "$SMPROGRAMS\VALENTINA.lnk" "$INSTDIR\VALENTINA.exe" "serve --open-browser" "$INSTDIR\VALENTINA.exe" 0
  WriteRegStr HKCU "${REGKEY}" "DisplayName" "VALENTINA"
  WriteRegStr HKCU "${REGKEY}" "DisplayVersion" "${PRODUCT_VERSION}"
  WriteRegStr HKCU "${REGKEY}" "InstallLocation" "$INSTDIR"
  WriteRegStr HKCU "${REGKEY}" "DisplayIcon" "$INSTDIR\VALENTINA.exe"
  WriteRegStr HKCU "${REGKEY}" "UninstallString" '$\"$INSTDIR\Uninstall.exe$\"'
  WriteRegStr HKCU "${REGKEY}" "QuietUninstallString" '$\"$INSTDIR\Uninstall.exe$\" /S'
  WriteRegDWORD HKCU "${REGKEY}" "NoModify" 1
  WriteRegDWORD HKCU "${REGKEY}" "NoRepair" 1
  IfErrors install_failed
  Goto install_done
  install_failed:
    MessageBox MB_OK|MB_ICONSTOP "Установка не завершена: файл или регистрация недоступны. Закройте приложение и повторите установку." /SD IDOK
    SetErrorLevel 4
    Quit
  install_done:
SectionEnd

Function un.onInit
  !insertmacro Initialize
  !insertmacro CheckPaths
  ReadRegStr $0 HKCU "${REGKEY}" "InstallLocation"
  ${If} $0 != $INSTDIR
    SetErrorLevel 3
    Quit
  ${EndIf}
  !insertmacro CheckUnlocked
FunctionEnd

Section "Uninstall"
  ClearErrors
  Delete "$INSTDIR\VALENTINA.exe"
  IfErrors uninstall_failed
  Delete "$INSTDIR\README.txt"
  Delete "$INSTDIR\VALENTINA-BUILD-MANIFEST.json"
  Delete "$INSTDIR\SHA256SUMS.txt"
  Delete "$INSTDIR\LICENSES\LICENSE.huginn.txt"
  Delete "$INSTDIR\LICENSES\THIRD-PARTY-NOTICES.txt"
  IfErrors uninstall_failed
  RMDir "$INSTDIR\LICENSES"
  Delete "$SMPROGRAMS\VALENTINA.lnk"
  DeleteRegKey HKCU "${REGKEY}"
  Delete "$INSTDIR\Uninstall.exe"
  ; Non-recursive: unknown files and ALL application data survive.
  RMDir "$INSTDIR"
  Goto uninstall_done
  uninstall_failed:
    SetErrorLevel 4
    Quit
  uninstall_done:
SectionEnd
