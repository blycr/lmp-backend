param(
  [string]$TaskName = "LmpServer"
)

Unregister-ScheduledTask -TaskName $TaskName -Confirm:$false -ErrorAction SilentlyContinue

