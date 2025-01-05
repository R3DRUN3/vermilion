# vermilion  

[![CI](https://img.shields.io/github/actions/workflow/status/R3DRUN3/vermilion/ci.yml?label=CI)](https://github.com/R3DRUN3/vermilion/actions/workflows/ci.yml)  [![License: Unlicense](https://img.shields.io/badge/license-Unlicense-blue.svg)](http://unlicense.org/)  ![Red Team Badge](https://img.shields.io/badge/Team-Red-red) [![Go Report Card](https://goreportcard.com/badge/github.com/r3drun3/vermilion)](https://goreportcard.com/report/github.com/r3drun3/vermilion)  

<img src="./docs/media/vermilion_logo.png" width="250x" />  


Linux post exploitation tool for info gathering and exfiltration 🐧 📡 💀



## Abstract  
`Vermilion` is a simple and lightweight CLI tool designed for rapid collection, and optional exfiltration of sensitive information from Linux systems.  
Its primary purpose is to streamline the process of gathering critical data in red teaming scenarios.  


> [!CAUTION]  
> This tool is in the early stages of development; as such, it may contain bugs or unhandled edge cases.    
> Vermilion has been designed as a resource for red teaming campaigns and/or educational purposes.  
> The author assumes no responsibility for the weaponization of this tool or the improper handling of sensitive data collected through its use.  


## How It Works  
Vermilion collects system information, env vars and sensitive directories/files, such as `.ssh`, `.bash/zsh_history`, `.aws`, `.docker`, `.kube`, `.azure`, `/etc/passwd`, `/etc/group` (and many others), then creates a compressed archive containing them.  

Additionally, it provides the option to exfiltrate the collected data via an HTTP `POST` request to a specified endpoint.   

The implementation of the endpoint for exfiltration is outside the scope of this tool; for an example, refer to [*this*](https://github.com/R3DRUN3/sploitcraft/tree/main/red-team-infra#deploy-a-lambda-function-for-data-exfiltration) resource.


## Example Use Case   
Imagine being engaged in a red teaming campaign and successfully compromising a Linux machine.    
Linux environments often are treasure trove of sensitive data and information due to their use as servers and their integration with other systems and softwares.  
Therefore, it is crucial to have an automated tool that enables rapid collection and exfiltration of sensitive information, such as environment variables and strategic directories, within seconds.  

This is where *Vermilion* proves helpful!  

### Video Demo


https://github.com/user-attachments/assets/76d510fe-2aac-4014-b3d6-c6b5563aa057



## Getting Started  

In order to get started with vermilion, follow the [*docs*](./docs/README.md).  

