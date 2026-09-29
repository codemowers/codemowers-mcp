# Developing PHP applications

Follow `building/general` for the single-process contract and legacy exceptions. PHP-FPM's master/worker model, even with `pm.max_children = 1`, does not meet that contract for new services; do not bundle it with nginx and a supervisor. See [PHP-FPM configuration](https://www.php.net/manual/en/install.fpm.configuration.php).

## Example: Framework X

[Framework X](https://framework-x.org/docs/getting-started/quickstart/) illustrates a long-lived PHP service using its ReactPHP server and direct `php public/index.php` entrypoint, rather than FPM integration. Choose a framework to fit the application.

Keep I/O asynchronous throughout the request path; blocking libraries still block the event loop. Use the process's shared metrics registry and bounded shutdown described in `developing/metrics` and `developing/general`.
