param(
  [string]$TaskName = "LmpServer",
  [string]$Executable,
  [string]$WorkingDir
)

if (-not (Test-Path -Path $Executable)) {
  throw "Executable not found: $Executable"
}
if (-not (Test-Path -Path $WorkingDir)) {
  throw "Working directory not found: $WorkingDir"
}

$action = New-ScheduledTaskAction -Execute $Executable -WorkingDirectory $WorkingDir
$trigger = New-ScheduledTaskTrigger -AtStartup
$settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -RestartCount 3 -RestartInterval (New-TimeSpan -Minutes 1)

Register-ScheduledTask -TaskName $TaskName -Action $action -Trigger $trigger -Settings $settings -RunLevel Highest -User "SYSTEM" -Force

