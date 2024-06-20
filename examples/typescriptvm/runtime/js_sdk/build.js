#!/usr/bin/env node
const { exec } = require('child_process');
const path = require('path');
const fs = require('fs');
// Extract the source file and output WASM file paths from the command line arguments
let [sourceFile, wasmOutputFile] = process.argv.slice(2);

if (!sourceFile) {
    console.error('Usage: `node build.js <source_file_path> <optional_wasm_output_file_path>`');
    process.exit(1);
}

// Define the temporary JS output file path (used as intermediary step)
wasmOutputFile = wasmOutputFile || sourceFile.replace('.ts', '.wasm');
const jsOutputFile = wasmOutputFile.replace('.wasm', '.temp.js');

//create wasmOutputFile dir if not exist
const outputDir = path.dirname(wasmOutputFile);
if (!fs.existsSync(outputDir)) {
    fs.mkdirSync(outputDir, { recursive: true });
}

// Build the JavaScript bundle with esbuild
//FIXME: return --minify
exec(`npx esbuild ${sourceFile} --bundle --outfile=${jsOutputFile}`, (err, stdout, stderr) => {
    if (err) {
        console.error(`Error during JS build: ${stderr}`);
        process.exit(1);
    }
    console.log(stdout);

    // Compile the JavaScript bundle to WASM with javy-cli

    getJavyBinaryPath().then(javyBinaryPath => {
        exec(`${javyBinaryPath} compile -d ${jsOutputFile} -o ${wasmOutputFile}`, (err, stdout, stderr) => {
            if (err) {
                console.error(`Error during WASM build: ${stderr}`);
                process.exit(1);
            }
            console.log(stdout);
            console.log(`Build successful: ${wasmOutputFile}`);
        })
    }).catch(e => {
        console.error(e)
        process.exit(1)
    })
});

async function getJavyBinaryPath() {
    const JAVY_VERSION = "v3.0.0"
    const javyBinaryUrl = `https://github.com/bytecodealliance/javy/releases/download/${JAVY_VERSION}/javy-${platarch()}-${JAVY_VERSION}.gz`;
    const binaryPath = path.join(cacheDir(), javyBinaryUrl.split("/").pop().replace(".gz", ""));

    if (!fs.existsSync(binaryPath)) {
        console.log("Downloading javy binary")

        await downloadGzBinary(javyBinaryUrl, binaryPath)
    }

    return binaryPath
}


function cacheDir(...suffixes) {
    const cachedir = require("cachedir")

    const cacheDir = path.join(cachedir("bincache"), ...suffixes);
    fs.mkdirSync(cacheDir, { recursive: true });
    return cacheDir;
}

function binaryPath(version) {
    return path.join(cacheDir(), `${NAME}-${version}`);
}

function platarch() {
    let platform, arch;
    switch (process.platform.toLowerCase()) {
        case "darwin":
            platform = "macos";
            break;
        case "linux":
            platform = "linux";
            break;
        case "win32":
            platform = "windows";
            break;
        default:
            throw Error(`Unsupported platform ${process.platform}`);
    }
    switch (process.arch.toLowerCase()) {
        case "arm":
        case "arm64":
            arch = "arm";
            break;
        // A 32 bit arch likely needs that someone has 32bit Node installed on a
        // 64 bit system, and wasmtime doesn't support 32bit anyway.
        case "ia32":
        case "x64":
            arch = "x86_64";
            break;
        default:
            throw Error(`Unsupported architecture ${process.arch}`);
    }
    const result = `${arch}-${platform}`;
    const SUPPORTED_TARGETS = [
        "arm-linux",
        "arm-macos",
        "x86_64-macos",
        "x86_64-windows",
        "x86_64-linux",
    ];

    if (!SUPPORTED_TARGETS.includes(result)) {
        throw Error(
            `Unsupported platform/architecture combination ${platform}/${arch}`
        );
    }
    return result;
}

const gunzip = require('zlib').createGunzip()

async function downloadGzBinary(fromUrl, toPath) {
    const compressedStream = await new Promise(async (resolve, reject) => {
        console.log(`Downloading ${fromUrl} to ${toPath}`);
        const resp = await fetch(fromUrl);
        if (resp.status !== 200) {
            return reject(`Downloading ${fromUrl} failed with status code of ${resp.status}`);
        }
        resolve(resp.body);
    });
    const output = fs.createWriteStream(toPath);
    await new Promise((resolve, reject) => {
        require('stream').pipeline(compressedStream, gunzip, output, (err, val) => {
            if (err) return reject(err);
            return resolve(val);
        });
    });

    await fs.promises.chmod(toPath, 0o775);
}