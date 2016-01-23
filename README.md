# pipethis

(I would have called it `stop-piping-the-internet-into-your-shell`, but that
seemed too long.)

## `tl;dr`

Instead of

```
$ curl -sSL https://get.rvm.io | bash
```

do

```
$ pipethis https://get.rvm.io
```

or (assuming not everybody has adopted my awesome idea but you still want to
improve your life)

```
$ pipethis --no-verify --inspect https://get.rvm.io
```

## Install

```
$ go get github.com/ellotheth/pipethis
```

or [download the binary](https://github.com/ellotheth/pipethis/releases) and
drop it in your `$PATH`.

or the easy way

```
curl https://raw.githubusercontent.com/ellotheth/pipethis/master/install.sh | bash
```
## Use

### People piping the installers

```
pipethis [ --target <exe> | --inspect | --editor <editor> | --no-verify ] <script>

--target <exe>

    The shell or other binary that will run the script. Defaults to the SHELL
    environment variable.

--inspect

    If set, open the script in an editor before checking the author.

--editor <editor>

    The editor binary to use when --inspect is set. Defaults to the EDITOR
    environment variable.

--no-verify

    If set, skips author and signature verification entirely. You'll need to
    set this if <script> doesn't support pipethis yet.
```

### People writing the installers

You can add one line to your installer script to make it support `pipethis`,
but there's other stuff to do as well:

1. Get an account on [Keybase](https://keybase.io). I know, Real Crypto Geeks™
   hate Keybase because Browser Crypto Is Unsafe™ and They Can Store
   Your Private Key®. It's a place to start, yo, just do it.
2. Add one line to your installation script to identify yourself. You can throw
   it in a comment:

    ```
    # // ; '' PIPETHIS_AUTHOR your_name_or_your_key_fingerprint
    ```

     1. If you don't want to store your signature at `<scriptname>.sig`, add
        another line to tell `pipethis` where you're storing it:
        
            # // ; '' PIPETHIS_SIG https://your.sig/location

3. Create a detached signature for the script. With Keybase, that's:

    ```
    $ keybase pgp sign -i yourscript.sh -d -o yourscript.sh.sig
    ```

   but you're a Real Crypto Geek™, so you'll use `gnupg`:

    ```
    $ gpg --detach-sign -a -o yourscript.sh.sig yourscript.sh
    ```

	Both those commands create ASCII-armored signatures. Binary signatures work
	too.
4. Pop both the script and the signature up on your web server.
5. Replace your copy-paste-able installation instructions!

## What's all this noise

Who's [piped](https://rvm.io/rvm/install) the
[installation script](https://github.com/npm/npm#fancy-install-unix) for their
[favorite tool](https://github.com/creationix/nvm#install-script) directly from
`curl` [into their shell](https://getcomposer.org/doc/00-intro.md#installation-linux-unix-osx)?
Show of hands? Come on, [you know](http://ipxe.org/)
[you have](https://docs.puppetlabs.com/pe/latest/install_agents.html#scenario-1-the-osarchitecture-of-the-puppet-master-and-the-agent-node-are-the-same).
Don't feel bad, so have I! So have we all, really. It's so easy, so fast, so
clean, so...well, **[bad](http://curlpipesh.tumblr.com/)
[for](https://jordaneldredge.com/blog/one-way-curl-pipe-sh-install-scripts-can-be-dangerous/)
[you](https://www.seancassidy.me/dont-pipe-to-your-shell.html)**.

- Network errors happen. Why pay for them in the middle of an install?
- Is your source served over SSL? No? Grats, you have no idea what you're
  downloading or where it came from. Exciting!
- What's that? Your source is served over SSL? Great! Any disgruntled employees
  have access to that server? Any trolls? Hey cool, you still have no idea what
  you're downloading!

There are simple solutions to some of those problems:

- Cache the script before you shove it into Bash.
- Use something like [vipe](https://joeyh.name/code/moreutils/) to pipe the
  script into an editor so you can review it before you run it.
- Use [hashpipe](https://github.com/jbenet/hashpipe) to check the file hash
  before you run it.

But simple solutions are, like, boring, and stuff.

## Why I'm here

The more interesting problem (to me, anyway) is authenticity. You trust whoever
wrote the script; how can you be reasonably sure the script you download is the
one they wrote?

### PGP. Clearly.

What if every installation script was embedded with the cryptographic signature
of its author, and you could verify the author and the script against the
signature when you ran it?

Enter `pipethis`.

## How it works

Scripts that support `pipethis` will have one or two special lines that
identify the script author, and optionally where the PGP signature lives.

```{.sh}
#!/usr/bin/bash

# PIPETHIS_AUTHOR gemma
# PIPETHIS_SIG    https://the.special.place/of/specialness

echo woooooo look how verified everything is!
```

`pipethis` checks [Keybase](https://keybase.io) for any users that match
`PIPETHIS_AUTHOR`. (It uses the same search you find in the search box on their
website, so you could use a username, a Twitter handle, or even a key
fingerprint.) It'll spit all the matches back at you in a list, along with all
their Keybase proofs. Once you choose one, `pipethis` grabs the public key for
that user. If you don't see the person you're looking for, you can bail. No
harm, no foul.

```
I found 2 results:

0:

     Identifier: gemma
        Twitter: ellotheth
         Github: ellotheth
    Hacker News: gemma
         Reddit: 
    Fingerprint: 417b9f99b7c04ccebd06777d0bc6bb965aa6f296
           Site: ramblinations.com
           Site: ramblinations.com


1:

     Identifier: gemmakbarlow
        Twitter: gemmakbarlow
         Github: gemmakbarlow
    Hacker News: gemmakbarlow
         Reddit: 
    Fingerprint: 1fd52e9237fef588e2d0d26100fee8d483374357
```

Once you've picked an author, `pipethis` will go grab their detached PGP
signature for the script. If `PIPETHIS_SIG` is not identified in the script,
`pipethis` will tack `.sig` onto the end of the script location and try that
instead.

With the signature and public key in hand, `pipethis` will verify that the
signature matches both the key and the script. If it does, you're good to go,
and `pipethis` will run the script for you (against the executable of your
choice). If not, `pipethis` dies, cleans itself up, and nobody ever has to know
that you almost pwned yourself.

## It's not done yet

`pipethis` works, but it can be better!

- If there were a non-interactive version, it could be inserted into a pipe
  chain like `curl -Ss http://pwn.me/please | pipethis | bash`. That'd be cool.
- There are zillions of other places to get public keys for people, and I want
  to support more of them. I think Keybase is stellar and I love what they're
  trying to do, but nobody likes to be locked in to one provider.
