
## Creating self-signed certificates

Generating CSRs

    openssl req -new -x509 -days 365 -nodes -out "server.crt" -newkey rsa:2048 -keyout "server.key"

The `-nodes` options specifies that the private key should _not_ be encrypted with a pass phrase.

Verify Certificats

    openssl x509 -text -noout -in server.crt  # view certificate entries
    openssl rsa -check -in server.key         # verify a private key

    openssl rsa  -noout -modulus -in server.key | openssl md5
    openssl x509 -noout -modulus -in server.crt | openssl md5

## Install fonts

    sudo apt install fonts-noto-color-emoji

Download [Hack.zip](https://github.com/ryanoasis/nerd-fonts/releases/tag/v2.1.0).

    cd ~/.local/share/fonts

Choose "Hack Nerd Font Mono Regular".
