# Developing Java applications

Apply `developing/general` for configuration, lifecycle and failures. Choose a framework to fit the application; Spring Boot is one example.

## Example: Spring Boot

- For a Spring Boot HTTP service, include `spring-boot-starter-web`, annotate the application with `@SpringBootApplication`, and start it with `SpringApplication.run`. Use `@RestController` and `@GetMapping` to define HTTP handlers.
- Add `spring-boot-starter-actuator` and `micrometer-registry-prometheus` for Prometheus metrics. Keep dependency versions compatible with the chosen Spring Boot release.
- Use the internal listener required by `developing/metrics`; expose only required management endpoints. Enable request latency histograms when needed.

For an application listening on port 8080, put this configuration in `src/main/resources/application.properties`:

```properties
server.port=8080
management.server.port=8081
management.endpoints.web.exposure.include=prometheus
management.metrics.distribution.percentiles-histogram.http.server.requests=true
```

This example exposes `/actuator/prometheus` on port 8081. Apply `developing/metrics` for metric safety and `deploying/observability` for the recorded scrape-transport choice.
