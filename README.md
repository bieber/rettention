# rettention

rettention is a command line tool to apply a retention policy to your
Reddit posts and comments.  The program executes one batch of
deletions in a single run, and it should be set up in a cron job to
maintain an ongoing retention policy.

In order to use rettention, you will need the ID and app secret of a
registered reddit app for API access.  You can set one up for yourself
at
[https://www.reddit.com/prefs/apps](https://www.reddit.com/prefs/apps).

## Command Line Flags

rettention expects a single command line argument, `-c` or
`--config-file`, indicating the location of a YAML config file.  If
not set, this will default to `rettention.yaml` in the current working
directory.

## Configuration File

Configure rettention with a YAML configuration file containing the
following values:

* **app_id**: The Reddit app ID.
* **app_secret**: The secret for your Reddit app.
* **serve_address**: The address at which to run the local server for
  OAuth.  For example use `:3000` to run on localhost at port 3000.
* **redirect_uri**: The redirect URI to use for OAuth.  If you used
  `:3000` for `serve_address` then the corresponding `redirect_uri`
  would be `http://localhost:3000`
* **credential_path**: The file to write Reddit OAuth credentials to.
* **users**: A map from user name to user configuration.

Each entry in the `users` map should have the username as its key and
the following values:

* **comment_duration**: The amount of time to retain comments by this
  user.  This must be either the string "forever" to turn off
  deletion, or a duration string that can be parsed by Go's
  [time.ParseDuration](https://pkg.go.dev/time#example-ParseDuration).
* **post_duration**: The amount of time to retain posts by this user.
  This must be either the string "forever" to turn off deletion, or a
  duration string that can be parsed by Go's
  [time.ParseDuration](https://pkg.go.dev/time#example-ParseDuration).

The auth command automatically adds entries to the `users` key as you
authenticate them.

## Commands

When you run rettention, you must specify one of two commands on the
command line.

### auth

The auth command launches a web browser to authenticate your reddit
account via OAuth.  As long as your `serve_address` and `redirect_uri`
configuration keys are set correctly, continuing through the OAuth
flow will generate credentials for your account.  After a successful
run, this command does two things:

* It adds the credentials for your account to the specified
  credentials file.
* If your username is not already present in the `users` key of the
  configuration file, it adds an entry with default values of
  "forever" for the retention periods.

### run

The run command runs a single round of content deletion.  It first
fetches the oldest comments and posts for each user specified in the
config and then begins deleting content older than the specified
retention period for its user and type.  It continues running until
all expired entries have been deleted.  This may take a while for
older accounts on the first run as Reddit's rate limits will slow the
script down.

## Credential Security

Note that rettention will write OAuth credentials in plain text to the
file specified in the `credential_path` configuration key.  Anyone
with access to this file will be able to read and write reddit content
as your user account, so be careful with access to it.  Ideally this
file should only be accessible to the user account that runs the
command.

## Disclaimer

The basic purpose of this program is to delete Reddit content.  If you
make a mistake in the configuration, or if there is a bug in the code,
it may delete content you didn't mean to.  This program comes with no
warranty and its author takes no responsibility for any erroneously
deleted data--**use at your own risk**.
