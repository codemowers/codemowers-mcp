# Building Java containers

Use the project's build tool and a compatible JVM runtime image; apply `building/general` for shared image conventions. A JVM application cannot run directly in `scratch`.

- Declare the Java version in build configuration and align builder/runtime versions. Choose maintained distributions, for example Temurin or Corretto.
- For Maven with Spring Boot, use `spring-boot-maven-plugin` to produce an executable JAR. Keep build-time checks enabled and match the copied filename to the actual artifact.

See `building/general` for the lolcatz Dockerfile examples; adapt the build and runtime stages to this language’s toolchain.
