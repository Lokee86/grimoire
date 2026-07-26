use std::env;

fn main() {
    let version = env::var("GRIMOIRE_RELEASE_VERSION").unwrap_or_else(|_| {
        env::var("CARGO_PKG_VERSION").expect("Cargo must provide CARGO_PKG_VERSION")
    });
    println!("cargo:rustc-env=GRIMOIRE_RELEASE_VERSION={version}");
    println!("cargo:rerun-if-env-changed=GRIMOIRE_RELEASE_VERSION");
}
