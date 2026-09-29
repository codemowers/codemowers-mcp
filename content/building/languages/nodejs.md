# Building Node.js containers

Apply `building/general` for image stages, locked dependencies, runtime identity and process ownership.

- Use the same pinned Node version across stages. Omit development-only packages from the runtime image.
- Launch Node directly with exec-form `CMD ["node", "server.js"]`, using the actual built entrypoint. Avoid production startup through `npm start`, `npm run` or `npx`: wrappers add process layers and may run hooks or resolve packages.
- Package-manager scripts remain suitable for installation, builds, checks and local development. A `start` script is not a defect when production launches Node directly.

See the [lolcatz Next.js frontend Dockerfile](https://github.com/codemowers/lolcatz/blob/main/frontend/Dockerfile) for a complete example.
