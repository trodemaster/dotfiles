<!-- Source: https://guide.macports.org/chunked/installing.html -->

# Chapter 2. Installing MacPorts

## Introduction

This chapter provides step-by-step instructions for installing MacPorts and its prerequisites. The Xcode installation section is specific to macOS. For other platforms, users should:
- Ensure a working C compiler is installed
- Follow the source installation instructions
- Complete the remaining chapter steps

## 2.1. Install Xcode

### Overview
[Xcode](https://developer.apple.com/xcode/) is an Apple package containing:
- Compilers
- Libraries
- Development tools for macOS applications

> Always install the latest Xcode version for your macOS release to prevent port install failures.

### Installation Instructions

#### For macOS 10.9 and Later
1. (Optional) Download Xcode from:
   - Apple developer website
   - Mac App Store
2. Open terminal
3. Run `xcode-select --install`
4. Click "Install"

#### For macOS 10.7-10.8
1. Download latest Xcode
2. In Xcode Preferences:
   - Go to Downloads section
   - Install Command Line Tools

#### For macOS 10.6
Two Xcode download options:
1. Xcode 3.2 (smaller download)
2. Xcode 3.2.6 with iOS SDK

Recommended installation selections:
- UNIX Development
- System Tools
- X11 SDK
- Command Line Support

#### For Older macOS Releases
Download appropriate Xcode version from Apple developer website, following similar selection guidelines.

## Note
Some ports require full Xcode; most do not. Users can choose minimal installation if needed.
