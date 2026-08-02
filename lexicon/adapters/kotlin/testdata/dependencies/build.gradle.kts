plugins {
    kotlin("jvm") version "2.0.21"
}

val dynamicVersion = "1.0"
val notDependency = "implementation(\"ignored:outside:1.0\")"

// implementation("ignored:comment:1.0")
dependencies {
    implementation("org.jetbrains.kotlin:kotlin-stdlib:2.0.21")
    api("com.example:public-api:1.0")
    api("com.example:public-api:1.0")
    compileOnly("org.jetbrains:annotations:24.1.0")
    runtimeOnly("com.example:runtime:3.0")
    testImplementation("org.junit.jupiter:junit-jupiter:5.11.0")
    kapt("com.google.dagger:dagger-compiler:2.52")
    ksp("com.google.devtools.ksp:symbol-processing-api:2.0.21-1.0.28")

    implementation(libs.kotlin.coroutines)
    implementation("com.example:dynamic:$dynamicVersion")
    implementation(project(":shared"))
    implementation(platform("com.example:bom:1.0"))
    implementation(files("libs/local.jar"))
    implementation(group = "com.example", name = "named", version = "1.0")
}
