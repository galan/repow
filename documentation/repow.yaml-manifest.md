# Repository manifest file (`repo.yaml`)

Repow encourages the concept of self-contained repositories by defining a `repo.yaml` manifest file, that contains meta-information about the repository content. These information are used to automatically update the repository at the hoster (eg. gitlab) with topics, description, configuration, etc.. It is optional to use and completly independent from the other commands described above.


## Commands
Commands utilizing the manifest-file:
* **validate**\
Validating the manifest-file (existince, patterns, usernames, etc.)
* **apply**\
Applying the manifest files values to the hoster repository
* **serve**\
Starts the webhook server, that will listen to the changes on the default branch to apply changes automatically on push events. You can configure slack to obtain notifications for invalid manifest files.


## repow.yaml
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

repow starts a webserver listening in port 8080 when called with the command `repow serve`. A ready-to-use docker-container is also [available](https://hub.docker.com/repository/docker/galan/repow).

### Configuration
You can utilize the configuration-file, or environment-variables (which is suggested started as container).

The list of environment variables that can be utilized:

* `REPOW_GITLAB_APITOKEN`
* `REPOW_GITLAB_HOST`
* `REPOW_GITLAB_SECRETTOKEN`
* `REPOW_OPTIONS_OPTIONALCONTACTS`
* `REPOW_SERVER_PORT`
* `REPOW_SLACK_APITOKEN`
* `REPOW_SLACK_CHANNELID`
* `REPOW_SLACK_PREFIX`
