use std::{path::Path, process::Command};

pub const BUILD_DIR_NAME: &str = "build";
const WASM_TARGET: &str = "wasm32-unknown-unknown";
const RELEASE_PROFILE: &str = "release";

#[allow(clippy::missing_panics_doc, clippy::module_name_repetitions)]
/// Put this in your build.rs file. It currently relies on `/build` directory to be in your crate root.
pub fn build_wasm_on_test() {
    println!("cargo:rerun-if-changed=build.rs");
    println!("cargo:rerun-if-changed=/build");

    // TODO:
    // remove these printlns
    let target = std::env::var("TARGET").unwrap();
    let profile = std::env::var("PROFILE").unwrap();

    if target != WASM_TARGET {
        let package_name = std::env::var("CARGO_PKG_NAME").unwrap();

        println!("cargo:warning=building `{package_name}` wasm file");

        let manifest_dir = std::env::var("CARGO_MANIFEST_DIR").unwrap();

        let profile = if profile == RELEASE_PROFILE {
            &profile
        } else {
            "test"
        };

        let target_dir = format!("{manifest_dir}/{BUILD_DIR_NAME}");

        let cargo_build_output = Command::new("cargo")
            .arg("build")
            .arg("--target")
            .arg(WASM_TARGET)
            .arg("--profile")
            .arg(profile)
            .arg("--target-dir")
            .arg(&target_dir)
            .output()
            .expect("command should execute even if it fails");

        let profile = if profile == RELEASE_PROFILE {
            "release"
        } else {
            "debug"
        };

        let target_dir = Path::new(&target_dir)
            .join(WASM_TARGET)
            .join(profile)
            .join(format!("{package_name}.wasm"))
            .canonicalize()
            .expect("file was not in expected location");

        println!("cargo:warning=`.wasm` file at {target_dir:?}");

        if !cargo_build_output.status.success() {
            let stdout = String::from_utf8_lossy(&cargo_build_output.stdout);
            let stderr = String::from_utf8_lossy(&cargo_build_output.stderr);

            println!("cargo:warning=stdout:");

            for line in stdout.lines() {
                println!("cargo:warning={line}");
            }

            println!("cargo:warning=stderr:");

            for line in stderr.lines() {
                println!("cargo:warning={line}");
            }

            println!("cargo:warning=exit-status={}", cargo_build_output.status);
        }
    }
}