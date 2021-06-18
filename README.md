<p align="center">
  <img src="media/repow-02-512.png" width="128">
</p>

**repow** simplifies management of your local git repository zoo.

repow provids commands for everyday operations on the one hand, and automatically applying configuration via a manifest-file on the other. The commands can be used independently of each other.

Commands that target everyday operations:
* **clone** - clone multiple repos in parallel. Filter by topics (tags), patterns or starred.
* **cleanup** - non-destructive cleanup of remotely deleted or archived repositories.
* **update** - checks, fetches and pulls all of your local repositories in parallel and prints condensed commit messages.

Beside, repow encourages the concept of self-contained repositories by defining a `repo.yaml` manifest file, that contains meta-information about the repository content. These information are used to automatically update the repository at the hoster (eg. gitlab) with topics, description, configuration, etc..

Commands utilizing the manifest-file:
* **validate** - Validating the manifest-file (existince, patterns, usernames, etc.)
* **apply** - Applying the manifest files values to the hoster repository
* **serve** - Starts the webhook server, that will listen to the changes on the default branch to apply changes automatically on push events.


# Architecture and process

DIAGRAM

Example workflow:
1. The developer clones all the desired repositories, based on tags, path-patterns (includes/exludes), or starred ones using `repow clone`.
2. From time-to-time the developer cleans the directory, containing the git-repositories via `repow cleanup`. Projects that have been deleted at the hoster are moved into a subdirectory `_deleted`, archived ones into `_archived`.
3. Also in regular intervals the local repositories are fetched or pulled with `repow update fetch` respective `repow update pull`. An overview of the local repositories modifications can be obtained with `repow update check`.

If the manifest-file is used, the configuration can be applied using one of the following methods:
* For local execution using the `repow apply` command.
* For automatic execution by running the `repow serve` or the Docker-container, that listens to the hosters git push webhook. On event, the state of the manifest file is applied directly to the project. That way you keep even track of your project configuration.

Before applying, the `repo.yaml` file is validated. Then the following settings from the manifest-file are applied to the project:
* The project description is updated
* Topics (aka tags) are added
* Organization Units, languages and type are also added as topic
* Hoster-specific configurations are updated

With the help of these additional topics, cloning specific selections becomes much easier and more efficient.


# Commands
Detailed help for every command can be obtained using the commandline help via `repow help <command>`

# Repository manifest file `repo.yaml`

The manifest filename is `repo.yaml` and lives in the root of the git repository.

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

# Webhook/Docker
The webhook listens for push events on the default branch, and applies the manifest-file on events.

repow starts a webserver listening in port 8080 when called with the command `repow serve`. A ready-to-use docker-container exists here: https://hub.docker.com/repository/docker/galan/repow

# Hoster

## Gitlab
When using Gitlab for Git-hosting, you'll need to set the following environment Variable: `REPOW_GITLAB_API_TOKEN` (or alternatively `GITLAB_API_TOKEN`).

You define you token with API scope in your [Gitlab preferences](https://gitlab.com/-/profile/personal_access_tokens).

## Github
Currently not supported (planned). Help is of course welcome.
