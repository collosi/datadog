# Datadog Client

A basic "metric reporting" client library for Datadog (www.datadoghq.com), written in Go

## Design objectives
* Minimize latency between "counter updated" and "shows up on dashboard"
* Minimize expense of metric update call
* Minimize memory allocation

## Limitations:
* Only supports counters
* Can drop updates (probably doesn't matter, because they'll be included in the next update)
