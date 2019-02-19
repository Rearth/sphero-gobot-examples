# Sphero Examples
This repository contains a number of example programms to control a Sphero SPRK+ using GoBot, written in Go.\
Communication is done using Bluetooth Low Energy, so they need to run on a device that supports BLE. In linux, root permissions are required to use Bluetooth low energy, so you need to build the programs using 'go build' and then run the binary using 'sudo'