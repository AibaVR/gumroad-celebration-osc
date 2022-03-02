# Aibas Gumroad Celebration OSC for VRChat
The gumroad celebration client watches for new sales on your gumroad account and toggles a celebration parameter on your VRChat avatar using the OSC API. The parameter can be configured. When a sale is made it will set the parameter to `true` for 5 seconds before resetting it to `false`.

## Usage
- Get  the latest release at https://github.com/AibaVR/gumroad-celebration-osc/releases
- Unzip the archive
- Open the `env.local` file in a text editor (notepad will do)
- Head to https://app.gumroad.com/settings/advanced#application-form
- Create a new application (call it whatever you want)
- Click on the "Generate access token" button
- Copy your access token
- In the `env.local` file, paste the access token next to the `GUMROAD_ACCESS_TOKEN=` entry
- In the `env.local` file, enter the avatar parameter you want toggled next to the `VRC_PARAM=` entry
- Run the `.exe`
- In game ensure that OSC is enabled

## Compiling from source
### Requirements
- [GoLang](https://go.dev/dl/)

### Compiling
- Inside the project, install the dependencies with `go mod download`
- Copy `env.local.example` to `env.local` and fill out the env vars
- Run the project with `go run main.go` or build it with `go build`

## Contact
For feature requests or bug reports please either submit an issue here or message me on discord `ai#0001`
