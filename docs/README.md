# Vermilion Docs


## Installation  



## Debug  

You can use [*vscode*](https://code.visualstudio.com/) for debugging.  
this repo already contains the `.vscode/launch.json` file:  

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Debug vermilion with VsCode",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "${workspaceFolder}/main.go",
            "args": ["--noexf"],
            "env": {},
            "showLog": true,
            "cwd": "${workspaceFolder}"
        }
    ]
}
```  

Just modify this file to accomodate your configuration and launch the debug via vscode.  

## Test Exfiltration  

You can test data exfiltration locally by following [*these instructions*](../tests/exfiltration/README.md).  
  


