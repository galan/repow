<p align="center">
  <img src="media/repow-02-512.png" width="128">
</p>

**repow** simplifies management of your local git repository zoo.

repow provids commands for everyday operations on the one hand, and automatically applying configuration via a manifest-file on the other. The commands can be used completely independently of each other.

**The operations/commands can be used totally independent of each other!**
There is no need to have a `repo.yaml` file to use the `clone`, `update` or `cleanup` commands, and vice versa.

# Installation and setup

Either download the [binary](https://github.com/galan/repow/releases) directly, or simply use [mise](https://mise.jdx.dev): `mise use -g ubi:galan/repow`

To configure repow use a config file. The default path is `$home/user/.repow.yaml`. You can define a custom path by using the flag parameter `--config`. If neither a the default nor the custom file exists, repow will create a config with default values.

The config has the following structure
```yaml
repow:
  server:
    port: 8080
  optionalContacts: false 
gitlab:
  host: gitlab.com
  token: "your_gitlab_token"
  secretToken: "your_gitlab_secrettoken_here"
options:
  downloadRetryCount: 6
  style: flat
slack:
  token: "your_slack_token_here"
  channelId: "your_slack_channel_here"
  prefix: ":large_blue_circle:"
```

You can use environment variables to overide the keys above. To do so each sub section is separated by an underscore. E.g. you want to override repow.server.port the corresponding env variable is `REPOW_SERVER_PORT`.

The following configuration can be helpful to define default behaviour:
* `gitlab.token` - Required \
For Gitlab create a [Personal Access Token](https://gitlab.com/-/user_settings/personal_access_tokens) with API scope.
* `gitlab.host`\
Set if you use a self-host Gitlab
* `options.style`
Define your default clone style: flat or recursive (preserve groups)

# Commands

First - don't panic! Detailed help for every command can be obtained at every point using the commandline help via `--help`

Commands that target everyday operations:
* **clone**\
Clones multiple repos in parallel. Filter by topics (tags), patterns or starred. Either clones the repositories flat, or recursive using the group structure as directories.
* **update**\
checks, fetches and pulls all of your local repositories in parallel and prints condensed commit messages.
* **cleanup**\
Non-destructive cleanup of remotely deleted or archived repositories.

Beside, repow encourages the concept of self-contained repositories by defining a `repo.yaml` manifest file, that contains meta-information about the repository content. These information are used to automatically update the repository at the hoster (eg. gitlab) with topics, description, configuration, etc..

Commands utilizing the manifest-file (completely optional):
* **validate**\
Validating the manifest-file (existince, patterns, usernames, etc.)
* **apply**\
Applying the manifest files values to the hoster repository
* **serve**\
Starts the webhook server, that will listen to the changes on the default branch to apply changes automatically on push events. You can configure slack to obtain notifications for invalid manifest files.


# Architecture and process

Example developer routine:

1. The developer clones all the desired repositories, based on tags, path-patterns (includes/exludes), or starred ones using `repow clone`. Executing this command again will only clone newly added repos.
2. From time-to-time the developer cleans the directory, containing the git-repositories via `repow cleanup`. Projects that have been deleted at the hoster are moved into a subdirectory `_deleted`, archived ones into `_archived`.
3. Also in regular intervals the local repositories are fetched or pulled with `repow update fetch` respective `repow update pull`. An overview of the local repositories modifications can be obtained with `repow update check`.

If the manifest-file is used (not necessary for the steps above and completly optional), the configuration can be applied using one of the following methods:
* For local execution using the `repow apply` command.
* For automatic execution by running the `repow serve` or the Docker-container, that listens to the hosters git push webhook. On event, the state of the manifest file is applied directly to the project. That way you keep even track of your project configuration.

Before applying, the `repo.yaml` file is validated. Then the following settings from the manifest-file are applied to the project:
* The project description is updated
* Topics (aka tags) are added
* Organization Units, languages and type are also added as topic
* Hoster-specific configurations are updated

With the help of these additional topics, cloning specific selections becomes much easier and more efficient.


# Repository manifest file (`repo.yaml`)

If the usage of the repository-metadata is wanted, there are some conventions and fields that can be used. The manifest filename is `repo.yaml` and lives in the root of the git repository.

Example file:
```yaml
name: my-project
description: Lorem Ipsum
type: service
languages: [java, kotlin]
topics: [foo, bar]
org:
  chapter: backend
  squad: user
annotations:
  deprecated: "false"
  acme.corp/access: "iam"
contacts:
  - galan
gitlab:
  wiki_access_level: "private"
  shared_runners_enabled: false
  forking_access_level: false
  only_allow_merge_if_pipeline_succeeds: true
  remove_source_branch_after_merge: true
```

* `name`: needs to be the same as the projects name
* `description`: Short precice description of the projects purpose
* `type`: A freely definable value for a type. Could be any string, recommended are values such as "service", "library", "tooling", ...
* `languages`: List of strings, containing one-letter words for the languages used by this project
* `topics`: List of freely definable topics names. Obey the pattern `a-z0-9-`.
* `org`: Map of freely definable key/value-pairs, that associate the project to your organization landscape.
* `annotations`: A freely definable key/value-structure for your own metadata (influenced by kubernetes annotations)
* `contacts`: List of users, that are associated with the project. How this is used depends on your organization-structure. Eg. it can be used to give other developers go-to persons for questions, merge-requests, etc..
* `gitlab`: Provides several gitlab hoster-specific project settings, that can be modified. The following values are supported at the moment: `wiki_access_level`, `issues_access_level`, `forking_access_level`, `build_timeout`, `only_allow_merge_if_pipeline_succeeds`, `only_allow_merge_if_all_discussions_are_resolved`, `remove_source_branch_after_merge`, `shared_runners_enabled`. If you miss a settings, feel free to open an issue.

The example above will result in the following topics: `language_java`, `language_kotlin`, `foo`, `bar`, `org_chapter_backend`, `org_squad_user`

### Webhook/Docker
The webhook listens for push events on the default branch, and applies the manifest-file on events.

repow starts a webserver listening in port 8080 when called with the command `repow serve`. A ready-to-use docker-container exists here: https://hub.docker.com/repository/docker/galan/repow
