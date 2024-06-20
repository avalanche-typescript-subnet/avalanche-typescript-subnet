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
    exec(`npx javy-cli compile -d ${jsOutputFile} -o ${wasmOutputFile}`, (err, stdout, stderr) => {
        if (err) {
            console.error(`Error during WASM build: ${stderr}`);
            process.exit(1);
        }
        console.log(stdout);
        console.log(`Build successful: ${wasmOutputFile}`);
    });
});