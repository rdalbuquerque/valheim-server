param (
    [ValidateSet("Deploy", "Debug")]
    [string]$To
)

$hostfiles = Get-ChildItem -Name "host*" 

$hostfilesmap = @{}

foreach ($file in $hostfiles) {
    $hostFileContent = Get-Content $file | ConvertFrom-Json
    if ($hostFileContent.customHandler.description.defaultExecutablePath -eq "dlv") {
        $hostfilesmap["debug"] = $file
    } else {
        $hostfilesmap["deploy"] = $file
    }
}

if ($To -eq "Debug") {
    Rename-Item $hostfilesmap["deploy"] "host.prod.json"
    Rename-Item $hostfilesmap["debug"] "host.json"
} else {
    Rename-Item $hostfilesmap["debug"] "host.debug.json"
    Rename-Item $hostfilesmap["deploy"] "host.json"
}

