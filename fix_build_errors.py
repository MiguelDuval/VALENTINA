path = 'C:/Users/migue/Projects/VALENTINA/scripts/build-valentina.ps1'
with open(path, 'r', encoding='utf-8') as f:
    content = f.read()

# Add helper function after param block
helper = '''
function Throw-Error {
    param([string]$Message)
    Write-Host "BUILD_ERROR: $Message" -ForegroundColor Red
    exit 1
}

'''
content = content.replace(') $ErrorActionPreference = "Stop"', helper + ') $ErrorActionPreference = "Stop"')

# Replace throw statements
import re
content = re.sub(r'throw "([^"]+)"', r'Throw-Error "$1"', content)

with open(path, 'w', encoding='utf-8') as f:
    f.write(content)
print('Build script updated with Throw-Error helper')
