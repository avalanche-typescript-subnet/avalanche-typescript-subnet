#!/bin/sh

( [ $# -eq 0 ] || [ $# -gt 2 ] ) && echo "Usage: $0 [ts-file] [optional wasm-file]" && exit 1

_tsfile=$1
_jsfile=${_tsfile%.*}.temp.js
_wasmfile=${2:- ${_tsfile%.*}.wasm}

npx esbuild ${_tsfile} --bundle --outfile=${_jsfile} --target=es2020 --format=esm
#convert path to absolute
tsfile=$(readlink -e $_tsfile)
[ $? -ne 0 ] && echo "Could not find \"$_tsfile\"" && exit 1

temp_js_name=$(mktemp -p ./)
temp_wasm_name=$(mktemp -p ./)

cp $_jsfile $temp_js_name

docker run --rm -v ./:/out javy-callback:latest compile -d /out/$temp_js_name -o /out/$temp_wasm_name

cp $temp_wasm_name $_wasmfile

rm -f $temp_js_name
rm -f $temp_wasm_name
rm -f ${HOME}/.cache/*.cwasm