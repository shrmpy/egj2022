# egj2022
Let's finish something!
 Magnet theme [Ebitengine Game Jam 2022](https://itch.io/jam/ebiten-game-jam)

When done 🤞, download will be on the [itch.io site here](https://shrmpy.itch.io/egj2022)


## Quickstart
```bash
git clone https://github.com/shrmpy/egj2022
cd egj2022 && go build 
./egj2022
```
## Build in Local Container
```bash
cd egj2022
docker build -t bc .
docker run -ti --rm --entrypoint sh -v $PWD:/opt/test bc
go build -o test
cp test /opt/test/egj2022
exit
./egj2022
```
## Make your own snap package
[![egj2022](https://snapcraft.io/egj2022/badge.svg)](https://snapcraft.io/egj2022)
```bash
# ub server includes a empty lxd?
sudo snap remove --purge lxd
# reinstall lxd
sudo snap install lxd
sudo lxd init --auto
sudo usermod -a -G lxd ${USER}
# view config
lxc version
lxc profile show default
lxc storage show default
echo 'export SNAPCRAFT_BUILD_ENVIRONMENT=lxd' >> ~/.profile
sudo reboot
# retrieve YAML 
git clone https://github.com/shrmpy/egj2022.git
cd egj2022
# make snap 
snapcraft
# local install
sudo snap install egj2022_0.0.1_arm64.snap --dangerous
# start program
egj2022
```


## Credits

Github workflow
 by [Siôn le Roux](https://github.com/sinisterstuf/ebiten-game-template) ([LICENSE](https://github.com/sinisterstuf/ebiten-game-template/blob/main/LICENSE))

Font Renderer
 by [tinne26](https://github.com/tinne26/etxt)
 ([LICENSE](https://github.com/tinne26/etxt/blob/main/LICENSE))

Ebitengine
 by [Hajime Hoshi](https://github.com/hajimehoshi/ebiten/)
 ([LICENSE](https://github.com/hajimehoshi/ebiten/blob/main/LICENSE))

DejaVu Sans Mono
 by [DejaVu](https://dejavu-fonts.github.io/)
 ([LICENSE](https://github.com/dejavu-fonts/dejavu-fonts/blob/master/LICENSE))

