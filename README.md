# Haul Equalizer (A fancy name for a load balancer)

This project aims to build a simple load balancer from scratch based in the popular ones like NGINX for example. I used a simple algorithm to check the servers health and handle the load accordingly. I know that the code is very wonky but the main purpose was to learn how load balancers work behind curtains.

# Features
- A simple round-robin algorithm to balance the loads.
- A server health checkup function for availability.

# Running locally
Just run the main.go file to start the load balancers along with the debug servers.

```bash
    go run main.go
```

PS: The debug servers are hardcoded to run on ports 9001 to 9003, If you are using those ports, you can change it in the backendServers variable

If the program did not exploded, just test if the load balancer is working.

```bash
    curl http://localhost:9000
```
If everything works successfully you should receive a message from a healthy server.
