<p align="center">
  <img src="media/repow-02-512.png" width="128">
</p>

**repow** simplifies management of your local git repository-zoo.

Some things you can do with repow:

* ⬇️ Clone all (or a filtered set) of your git-repositories into a directory (flat or recursive)
* ✨ Update all of your git-repositories recursivly (only check, fetch or pull fast-forward)
* 🧹 Cleanup all of your removed or archived git-repositories recursivly

Repow is simple, fast and easy to set up. It comes with several commands for everyday operations to ease the burden of keeping the cloned repos up-to-date.

Besides it also offers applying configuration and meta-data via a manifest-file (completly optional and independent).



# Installation and setup

Either download the [binary](https://github.com/galan/repow/releases) directly, or simply use [mise](https://mise.jdx.dev): `mise use -g ubi:galan/repow`

Next up is configuration. You specify most settings via configuration-file or environment-variables. For more details on the configuraiton options, see the configuration section below. To get things started, use the following environment-variables:

* `REPOW_GITLAB_APITOKEN` - Required for Gitlab, create a [Personal Access Token](https://gitlab.com/-/user_settings/personal_access_tokens) with API scope.
* `REPOW_GITLAB_HOST` - Set this, if you use a self-hosted Gitlab instance.
* `REPOW_OPTIONS_STYLE` - Define your default clone style (`flat` (default) or `recursive`).



# Commands
First - Don't panic! Detailed help for every command can be obtained at every point by omitting the arguments or via `--help`.

Commands that help you on everyday operations keeping pace with the growing amount of repositories:

### ⬇️ clone
Clones multiple repos in parallel. Filter by topics (tags), patterns (include and exclude) or starred favorites. Either clones the repositories `flat` (default), or `recursive` using the group structure as directories. You should be aware, that mixing both modes in the same target directory will result in mixed repository layouts. Also you have to keep the repository-names unique when using `flat`.

You can and should repeat this as often as you like, as only repositories will be cloned, that are not locally cloned yet.

Examples
```bash
# Clones everything you have access to into the current directory
repow clone . 

# Clones everything that matches the include filters group path (multiple possible)
repow clone . -i "^my-group/sub" -i "^other.*/regex-[0-9]{0-9}" -e "^private/"

# Clones everything that contains the topics aka labels
repow clone . -t "library" -t "a-team"

# Combination of all above is also possible
repow clone . -e "^private/" -t "library"
```


### ✨ update
This checks, fetches and pulls all of your local repositories in parallel and prints condensed commit messages. Hint: Use `-q` to hide untouched repositories in the output.

Examples
```bash
# Lists all repositories, that contain changes, and their current branch
repow update check . -q

# Fetches changes for all repositories for the current branch, and prints only those with changes
repow update fetch . -q

# If fast-forward is possible, pulls changes for all repositories for the current branch, and prints only those with changes
repow update pull . -q
```


### 🧹 cleanup
Non-destructive cleanup of remotely deleted or archived repositories. Those repositories will be moved into a subdirectory.

Examples
```bash
# Checks all repositories if they are removed or archived, and moves them non-destructive aside
repow cleanup . -q
```


# Configuration

repow uses the following configuration presedence (last will overwrite previous):
* default values
* configuration-file
* environment-variables
* command-line flags

The exact configuration-file location depends on your OS, but is printed in the general help by just writing `repow`.

You can overwrite the location using the `-c <location>` flag. This is the structure (including the default values) for all possible settings:

```yaml
options:
  style: flat
  parallelism: 32
  optionalcontacts: false
server:
  port: 8080
gitlab:
  host: gitlab.com
  apitoken:
  downloadretrycount: 6
  secrettoken:
slack:
  token:
  channelid:
  prefix: ":repow:"
```

Environment-variables use the same structure, but start with `REPOW_` followed by the uppercase, snakecased setting. As example, the style can be set via `REPOW_OPTIONS_STYLE`, the gitlab apitoken via `REPOW_GITLAB_APITOKEN`.


# Performance
repow aims to be simple to use and fast using parallelization.

Some relative benchmarks from my machine to get an idea. Those will vary depending on the machine-type, host, network, repository-sizes:

real-world git selfhosted instance:
* Clone 515 repositories: 2 min 17 sec
* Clone 317 repositories: 1 min 25 sec 
* Fetch 631 repositories: 23 sec
* Pull 631 repositories: 25 sec
* Cleanup 631 repositories: 6,8 sec


# Bonus: manifest-file

Beside, repow encourages the concept of self-contained repositories by defining a `repo.yaml` manifest file, that contains meta-information about the repository content. These information are used to automatically update the repository at the hoster (eg. gitlab) with topics, description, configuration, etc.. It is optional to use and completly independent from the other commands described above.

Commands utilizing the manifest-file (optional):
* **validate** - Validating the manifest-file (existince, patterns, usernames, etc.)
* **apply** - Applying the manifest files values to the hoster repository
* **serve** - Starts the webhook server, that will listen to the changes on the default branch to apply changes automatically on push events. You can configure slack to obtain notifications for invalid manifest files.

Read the [manifest-file](https://github.com/galan/repow/blob/master/documentation/repow.yaml-manifest.md) article, to get more insights about the possibilities.
