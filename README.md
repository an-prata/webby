# webby
Web server with some logging and whatever else I end up adding. To use just put all your website into the `/srv/webby/website` directory, for each directory located within if it contains an `index.html` file that will be served for requests to the directory, e.g. `https://an-prata.it/` serves the file `/srv/webby/website/index.html`. For serving HTTPS requests you can add paths to your certificate and key files in the config.

## Installing
I maintain an AUR package for webby:
```
yay -S webby-git
```

## webby.service
In the root of this repo there is a unit file named `webby.service`. If you install the AUR package this gets moved to `/usr/lib/systemd/system/`. If you do not install from the AUR you should move this file to `/etc/systemd/system/`, see https://wiki.archlinux.org/title/systemd#Writing_unit_files.

## Configuring
Basic configuration can be done with the `/etc/webby/config.json` file. If this file is absent `webby` will use a default configuration. The default configuration is located at the root of this repo and will be auto populated upon an AUR installation.
