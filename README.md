# spotify-servant

## What

Simple script to archive the *Discover Weekly* playlist.

## How

It is a CLI script for now, might change it later. It caches the access token.

It doesn't add duplicate songs and it adds the new songs in the beginning.

## Why

Last week I found a couple of really nice songs in the weekly playlist,
but I forgot to save them. Now the are lost :(

This script should make sure I don't lose any more songs this way.


## SETUP

I use `mkcert` to create local certificates that is supported.

Also need to google cloud components.

Then you run the command:

```shell
CLOUDSDK_PYTHON=python2 dev_appserver.py --support_datastore_emulator=yes app.yaml batch.yaml --clear_datastore=false --datastore_consistency_policy=consistent --env_var 'GOOGLE_APPLICATION_CREDENTIALS'=<credentials-file.json> --ssl_certificate_path <cert> --ssl_certificate_key_path <key>
```

Fill in the values in the `app.yaml` environment variables.
