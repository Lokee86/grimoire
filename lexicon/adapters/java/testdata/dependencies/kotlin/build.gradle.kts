plugins {
    kotlin("jvm") version "2.0.0"
}

dependencies {
    implementation("io.ktor:ktor-server-core:2.3.11")
    api("io.ktor:ktor-server-test-host:2.3.11")
    compileOnly("javax.servlet:javax.servlet-api:4.0.1")
    runtimeOnly("org.postgresql:postgresql:42.7.3")
    testImplementation("org.assertj:assertj-core:3.26.0")
    kapt("com.google.dagger:dagger-compiler:2.51.1")

    implementation(libs.ktor.server.core)
    implementation("org.example:dynamic:${property("version")}")
    implementation(project(":shared"))
}

// api("ignored:commented-kts:1.0")
val text = "runtimeOnly(\"ignored:string-kts:1.0\")"
