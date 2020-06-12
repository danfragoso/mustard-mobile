#!/bin/bash
adb shell pidof -s dev.fragoso.thdwb
while [ $? -ne 0 ]; do
  adb shell pidof -s dev.fragoso.thdwb
done
