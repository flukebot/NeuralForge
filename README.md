# README

## About

This is the official Wails React template.

You can configure the project by editing `wails.json`. More information about the project settings can be found
here: https://wails.io/docs/reference/project-config

## Live Development

To run in live development mode, run `wails dev` in the project directory. This will run a Vite development
server that will provide very fast hot reload of your frontend changes. If you want to develop in a browser
and have access to your Go methods, there is also a dev server that runs on http://localhost:34115. Connect
to this in your browser, and you can call your Go code from devtools.

## Building

To build a redistributable, production mode package, use `wails build`.


# Sound needs ffmpeg
Linux
```
sudo apt update
sudo apt install ffmpeg
```

Windows
On Windows:
Download the latest FFmpeg build from FFmpeg's official site.
Extract the downloaded file.
Add the bin directory (which contains ffmpeg.exe) to your system's PATH. To do this:
Search for "Environment Variables" in the Windows Start Menu.
Under "System Properties" -> "Advanced" -> "Environment Variables", find the "Path" variable under "System variables" and click "Edit".
Add the path to the bin directory of the extracted FFmpeg files.
Click "OK" to apply the changes.

# Verify Installation:
```
ffmpeg -version
```