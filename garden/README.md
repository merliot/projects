# Off-Grid Garden Watering System
This project is an off-grid garden watering system.

## Get
```
$ git clone https://github.com/merliot/projects.git
```

## Build
```
$ cd projects/garden
$ ./build
```
Builds the binary ~/go/bin/garden

## Run
```
$ cd projects/garden
$ ~/go/bin/garden
[dc_a6_32_7a_a6_d0] Merle version: v0.0.46
[dc_a6_32_7a_a6_d0] Model: "garden", Name: "eden"
[dc_a6_32_7a_a6_d0] Received [SYSTEM]: {"Msg":"_CmdInit"}
[dc_a6_32_7a_a6_d0] Basic HTTP Authentication enabled for user "merle"
[dc_a6_32_7a_a6_d0] Public HTTP server listening on port :80
[dc_a6_32_7a_a6_d0] Skipping public HTTPS server; port is zero
[dc_a6_32_7a_a6_d0] Private HTTP server listening on port :6000
[dc_a6_32_7a_a6_d0] Skipping tunnel to mother; missing host
[dc_a6_32_7a_a6_d0] Received [SYSTEM]: {"Msg":"_CmdRun"}
```
