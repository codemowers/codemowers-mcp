# Git usage

## History and integration

Use linear history for applications and microservices. Rebase feature work onto the target branch, then integrate with rebase or fast-forward; squash intermediate fixups when a request represents one logical change. This keeps changes easier to review, bisect and revert. See [Git SCM's rebasing guide](https://git-scm.com/book/en/v2/Git-Branching-Rebasing).

Allow ordinary merge commits only in large projects where several contributors actively change the same code sections and a designated integrator owns conflict resolution and review. Linux's [maintainer hierarchy](https://git-scm.com/book/en/v2/Distributed-Git-Distributed-Workflows) illustrates this model. Record the integrator and exception in the repository's contribution rules; project size alone does not justify it.

Review and test the final integrated revision as described in `building/ci`. A merge can contain edits absent from either parent, including legitimate conflict resolutions; Git calls this an [evil merge](https://git-scm.com/docs/user-manual#def_evil_merge). For a two-parent merge, use `git show --remerge-diff <commit>` to inspect changes from Git's reconstructed merge, or `git show --diff-merges=separate <commit>` to compare each parent. Combined diffs omit files that are unchanged relative to any parent. Rebase and squash results also need review.

Keep prototypes lightweight: commit intermediate work and clean it up as the design settles. Rewrite feature branches you own when nobody depends on their history, or when collaborators agree. Preserve published integration and release history that contributors, releases or automation depend on, and follow branch protections.

After rewriting a remote feature branch, use `git push --force-with-lease`. For a shared branch, verify its remote tip and use `--force-with-lease=<branch>:<expected-commit>` so background fetches cannot change the expectation. A lease prevents overwriting unexpected updates; coordinate with collaborators before rewriting shared work.

## Microservices and repository layout

Choose a monorepo or separate service repositories according to ownership, shared code and how often changes span services. The number of images does not dictate the number of repositories. Use linear integration history in either layout.

- **Monorepo:** A coherent change can update several services in one commit, and CI can build their images from that revision. A shared source commit does not make deployment atomic.
- **Separate repositories:** Each service has its own review, build and release history. Changes spanning services require coordinated commits and releases.

Record each image's source revision and digest, and record compatible image digests together for deployment. Follow `building/ci` for provenance and release records.

## Repository hosting configuration

Change repository settings to integrate pull or merge requests without merge commits. Local `pull.rebase` only controls your pulls; it does not configure the hosting service's merge button. Retain meaningful individual commits through rebase or fast-forward; allow squash when a request represents one logical change with intermediate fixups.

| Host | Repository settings for linear integration |
| --- | --- |
| GitHub | In **Settings → General → Pull Requests**, disable **Allow merge commits** and enable **Allow rebase merging**; optionally allow squash merging. Enable **Require linear history** in the target branch's ruleset or protection rule to reject merge commits pushed directly as well. See [rebase settings](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/configuring-pull-request-merges/configuring-commit-rebasing-for-pull-requests) and [linear history rules](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/available-rules-for-rulesets#require-linear-history). |
| GitLab | In **Settings → Merge requests → Merge method**, select **Fast-forward merge** and save. Rebase the source branch when it is behind the target. Configure **Squash commits when merging** according to commit quality. **Merge commit with semi-linear history** still creates merge commits and does not implement this policy. See [merge methods](https://docs.gitlab.com/user/project/merge_requests/methods/). |
| Gitea | Disable **Merge** and **Rebase then merge**, enable **Rebase** and make it the default. Optionally allow squash or fast-forward-only. |

Merge-method settings control PR integration through the host. Configure target-branch protections and direct-push permissions separately to prevent bypassing that workflow.

For Gitea, set the default for new repositories in `app.ini`:

```ini
[repository.pull-request]
DEFAULT_MERGE_STYLE = rebase
```

## User configuration

Set your author identity and make rebasing the default for pulls in `~/.gitconfig`. For example:

```ini
[user]
    email = lauri@codemowers.io
    name = Lauri Võsandi

[pull]
    rebase = true

[includeIf "gitdir:~/company-a/"]
    path = ~/company-a/gitconfig

[includeIf "gitdir:~/company-b/"]
    path = ~/company-b/gitconfig
```

The `[pull]` setting makes `git pull` rebase local commits onto upstream changes. Set it directly with `git config --global pull.rebase true`.

Use conditional includes for company-specific identities. The included files override the earlier `[user]` defaults for repositories under their matching directories. Inspect effective settings with `git config --show-origin --get <key>` when needed.
