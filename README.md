# webby
Web server with some logging and whatever else I end up adding. To use just put all your website into the `/srv/webby/website` directory, for each directory located within if it contains an `index.html` file that will be served for requests to the directory, e.g. `https://an-prata.it/` serves the file `/srv/webby/website/index.html`. For serving HTTPS requests place a certificate file into the `/srv/webby/` directory and name it `cert.pem`, do the same for a key file but name it `key.pem`.

## Installing
I maintain an AUR package for webby:
```
yay -S webby-git
```
