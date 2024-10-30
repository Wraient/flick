# Flick

Flick is a CLI tool for managing and watching TV shows and movies directly from the terminal or using Rofi for a GUI selection. Designed for simplicity and control, Flick allows seamless playback and tracking of episodes with customizable options.

## **Important** 
`Flick` uses [vadapav.mov](https://vadapav.mov) API and It also hosts the content. Please consider donating to [vadapav.mov](https://vadapav.mov)

Bitcoin 
```
bc1q8yxttnkuf3fygzxffj5wd85fgya0w39nr5tkau
```

Monero
```
42UXZfPr4SyZ4StxqNtA9HdvNr8ieSuMYdPxs3zL7qrKUnrWCMMUAuH9ARC342732VPS3KU6R8JbN15HWEdR234aPWF5ned
```
## Demo

CLI mode:

https://github.com/user-attachments/assets/c0578473-d463-41b9-aef5-346ecc1b05ce

Rofi mode:

https://github.com/user-attachments/assets/c751fdea-472d-4748-ba9e-b8d2a4885d07


## Features

- **Watch shows/movies**: Stream media directly in the CLI or through Rofi.
- **Playback management**: Continue, pause, and resume playback using MPV.
- **Progress tracking**: Automatically marks episodes as watched based on custom percentage.
- **Configurable storage and playback**: Set custom storage paths and MPV playback settings.
- **Customizable through CLI options**: Toggle between CLI and Rofi interface, configure playback speed, and more.

## Installing and Setup
> **Note**: `Flick` requires `mpv`, `rofi` for Video playback and Rofi support. These are included in the installation instructions below for each distribution.

### Linux
<details>
<summary>Arch Linux / Manjaro (AUR-based systems)</summary>

Using Yay

```
yay -Sy flick
```

or using Paru:

```
paru -Sy flick
```

Or, to manually clone and install:

```bash
git clone https://aur.archlinux.org/flick.git
cd flick
makepkg -si
sudo pacman -S rofi 
```
</details>

<details>
<summary> Debian / Ubuntu (and derivatives) </summary>

```bash
sudo apt update
sudo apt install mpv curl rofi
curl -Lo flick https://github.com/Wraient/flick/releases/latest/download/flick
chmod +x flick
sudo mv flick /usr/local/bin/
flick
```
</details>

<details>
<summary>Fedora Installation</summary>

```bash
sudo dnf update
sudo dnf install mpv curl rofi
curl -Lo flick https://github.com/Wraient/flick/releases/latest/download/flick
chmod +x flick
sudo mv flick /usr/local/bin/
flick
```
</details>

<details>
<summary>openSUSE Installation</summary>

```bash
sudo zypper refresh
sudo zypper install mpv curl rofi
curl -Lo flick https://github.com/Wraient/flick/releases/latest/download/flick
chmod +x flick
sudo mv flick /usr/local/bin/
flick
```
</details>

<details>
<summary>Generic Installation</summary>

```bash
# Install mpv, curl, rofi
curl -Lo flick https://github.com/Wraient/flick/releases/latest/download/flick
chmod +x flick
sudo mv flick /usr/local/bin/
flick
```
</details>

<details>
<summary>Uninstallation</summary>

```bash
sudo rm /usr/local/bin/flick
```

For AUR-based distributions:

```bash
yay -R flick
```
</details>

### [Windows Installer](https://github.com/Wraient/flick/releases/latest/download/FlickInstaller.exe)


## Usage

Run `flick -h` to see available commands and options:

Here's a table version for the command options:

| Option                          | Description                                                                          | Default                     |
|---------------------------------|--------------------------------------------------------------------------------------|-----------------------------|
| `-e`                            | Edit the Flick configuration file                                                    | N/A                         |
| `-next-episode-prompt`          | Prompt for the next episode playback (accepts true/false)                            | N/A                         |
| `-no-rofi`                      | Disable the Rofi interface; run in CLI mode                                          | N/A                         |
| `-percentage-to-mark-complete`  | Set the percentage of an episode to mark as complete                                 | `92`                        |
| `-player`                       | Set player for playback (only MPV supported)                                         | `"mpv"`                     |
| `-rofi`                         | Enable Rofi interface for selection                                                  | N/A                         |
| `-save-mpv-speed`               | Save MPV speed setting (accepts true/false)                                          | `true`                      |
| `-storage-path`                 | Define custom path for storage directory                                             | `$HOME/.local/share/flick`  |
| `-update`                       | Update the Flick script                                                              | N/A                         |

This makes it easy to scan and find each option’s purpose and default values!

### Examples

- **Watch with Rofi interface**:
  ```
  flick -rofi
  ```
- **Play next episode in CLI**:
  ```
  flick -next-episode-prompt
  ```
- **Change storage path**:
  ```
  flick -storage-path="/custom/path"
  ```

## Configuration

Edit the Flick configuration file to customize settings:
```
flick -e
```

## Dependencies
- mpv - Video player (vlc support might be added later)
- rofi - Selection menu

# API Used
- [vadapav.mov](https://vadapav.mov)

## License

Flick is open-source software licensed under the [MIT License](LICENSE).

# Credits
- [Lobster](https://github.com/justchokingaround/lobster) - For the motivation
