# Local development

- Use the project's Skaffold or Docker/Compose workflow for development, dependency management, builds and checks. Agents must not install application runtimes, packages or backing services on the host. Reuse existing host orchestration/editing tools; run installation inside build containers. Lockfile-only commands are acceptable when they install no host components.
- Use Compose for local development and integration testing. Match production's component layout, dependency versions and environment-variable names. Development images may add debugging tools and volume-mounted hot reload.
- Run Compose without TLS; IPv4-only local networking is acceptable. Keep image and Kubernetes listeners compatible with `deploying/networking`. Do not repeat defaults already supplied by the application or image.
- Keep local `.env` files untracked. Document startup, checks and deployment commands, including required configuration.
