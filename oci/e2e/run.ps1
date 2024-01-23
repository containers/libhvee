param(
    [Parameter(Mandatory,HelpMessage='folder on target host where assets are copied')]
    $targetFolder,
    [Parameter(Mandatory,HelpMessage='junit results filename')]
    $junitResultsFilename
)

# Run e2e
$env:PATH="$env:PATH;$env:HOME\$targetFolder;"
libhvee-e2e.exe --ginkgo.vv --ginkgo.junit-report="$targetFolder/$junitResultsFilename"
