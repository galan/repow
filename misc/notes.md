# Releasing
```
git tag -a v0.0.1 -m "release 0.0.1"
git push --tags
```

# Random thoughts
Old git code
```
		// root url: https://gitlab.com/...
		// * github.com > github host
		// * gitlab com > gitlab host. How about self-hosted?
		// api key

		//ccc, err := giturls.Parse("")
		//u, err := url.Parse("git@github.com:go-git/go-git.git")
		//u, err := url.Parse("https://github.com/go-git/go-git.git")
		//if err != nil {
		//	panic(err)
		//}
		//fmt.Println(u.Host)

		//sshIdFile := ssh_config.Get(u.Host, "IdentityFile")
		//cfg, err := ssh_config.Decode(strings.NewReader(config))
		//fmt.Println(cfg.Get("example.test", "Port"))
		//println(sshIdFile)

		//pem, _ := ioutil.ReadFile(pkeyfile)
		//signer, _ := ssh.ParsePrivateKey(pem)
		//aith := &ssh2.PublicKeys{User: "git", Signer: signer}

		//gitclient.Clone(dirReposRoot, "git@github.com:go-git/go-git.git")
		//gitclient.Clone(dirReposRoot, "git@gitlab.com:xxx/tests/project.git")
```

# Similar projects
* https://github.com/nosarthur/gita


# misc
tags:
- gitlab: a-z-_:
- github: a-z0-9- (Topics must start with a lowercase letter or number, consist of 35 characters or less, and can include hyphens.)

# TODOs
X argument handling
X gitlab token handling
X REPOW_GITLAB_API_TOKEN
X REPOW_GITHUB_API_TOKEN
X logo
X build-script/make/docker
X dockercontaintainer
X github repo
X github pipeline
X update readme
X update command docs
X brighten color project names
- static checks, eg "org_squad: xyz"
- query command
X improve "host" for hoster
