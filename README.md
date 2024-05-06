
# Stormfetch
## A simple linux fetch program written in go and bash

### Developers:
- [CapCreeperGR ](https://gitlab.com/CapCreeperGR)

### Project Information
Stormfetch is a program that can read your system's information and display it in the terminal along with ascii art of the linux distribution you are running

### Installation Guide
- Download the latest version of the plugin from this repository
- Run the following command to compile the project
```
make
```
- Run the following command to install stormfetch into your system. You may also append a DESTDIR variable at the end of this line if you wish to install in a different location
```
make install PREFIX=/usr SYSCONFDIR=/etc
```
