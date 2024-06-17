#!/bin/sh

printUsage () {
    cat << EOT
Usage: $0 compile [ts-file] [optional wasm-file]
       $0 emit-provider [wasm-file]

EOT
}

( [ $# -eq 0 ] || [ $# -gt 3 ] ) && printUsage && exit 1

case $1 in
  compile)
        shift 1
        _tsfile=$1

        [ -z "$_tsfile" ] && echo "Error: not specified a ts-file" && printUsage && exit 1
        readlink -e $_tsfile 
        [ $? -ne 0 ] && echo "Error: could not find \"$_tsfile\"" && exit 1

        _jsfile=${_tsfile%.*}.temp.js
        _wasmfile=${2:- ${_tsfile%.*}.wasm}

        echo "Compiling \"$_tsfile\" to \"$_jsfile\" and \"$_wasmfile\""

        npx esbuild ${_tsfile} --bundle --outfile=${_jsfile} --target=es2020 --format=esm
        #convert path to absolute
        tsfile=$(readlink -e $_tsfile)
        [ $? -ne 0 ] && echo "Could not find \"${_tsfile}\"" && exit 1

        temp_js_name=$(mktemp -p ./)
        temp_wasm_name=$(mktemp -p ./)

        cp ${_jsfile} ${temp_js_name}

        docker run --rm -v ./:/out javy-callback:latest compile -d /out/${temp_js_name} -o /out/${temp_wasm_name}

        cp ${temp_wasm_name} ${_wasmfile}
        rm -f ${temp_js_name}
        rm -f ${temp_wasm_name}
    ;;
  emit-provider)
        _wasmfile=$2

        [ -z $_wasmfile ] && echo "Error: missing a target wasm file" && printUsage && exit 1; 

        touch _wasmfile
        [ $? -ne 0 ] && echo "Could write to \"$_wasmfile\"" && exit 1

        path=$(readlink -e $_wasmfile)
        path=${path%/*}

        docker run --rm -v ${path}:/out javy-callback:latest emit-provider -o /out/javy_provider_1.4.0.wasm

        #clear cache
        rm -f ${HOME}/.cache/*.cwasm
    ;;
  *)
    echo "Unknown command: $1"
    printUsage
    exit 1
    ;;
esac

