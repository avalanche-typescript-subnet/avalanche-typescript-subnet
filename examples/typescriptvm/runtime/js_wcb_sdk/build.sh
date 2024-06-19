#!/bin/bash

set -e

CACHE_DIR=${HOME}/.cache/javy-callback
TEMP_DIR=/tmp
VERSION=v3.0.0-callback
OS=`uname -s`
OS=${OS,,}

ARCH=`uname -i`
ARCH=${ARCH,,}

[ ${ARCH} == 'aarch64' ] &&  ARCH="arm"

EXECUTABLE="javy-${ARCH}-${OS}-${VERSION}"
URL="https://github.com/crrrazyden/javy/releases/download/${VERSION}/javy-${ARCH}-${OS}-${VERSION}.gz"

printUsage () {
    cat << EOT
Usage: $0 compile [ts-file] [optional wasm-file]
       $0 emit-provider [wasm-file]

EOT
}

checkCache () {
    [ ! -d ${CACHE_DIR} ] && mkdir -p ${CACHE_DIR} && echo "need to download javy" && return
    [ ! -f ${CACHE_DIR}/${EXECUTABLE} ] && echo "need to download javy" && return
}

downloadJavy () {
  tmpfile=$(mktemp  ${TEMP_DIR}/tmp.XXXXXXXXXX.gz)
  curl -s -L -o ${tmpfile} ${URL}
  [ $? -ne 0 ] && echo "could not download javy" && return
  gzip -d ${tmpfile}
  [ $? -ne 0 ] && echo "could not download javy" && return
  mv ${tmpfile%.gz} ${CACHE_DIR}/${EXECUTABLE}
  chmod +x ${CACHE_DIR}/${EXECUTABLE}
  rm -f ${tmpfile%.gz}
}

getJavy () {
if [ "`checkCache`" = "need to download javy" ]; then
  echo "downloading javy"
  [ "`downloadJavy`" = "could not download javy" ] && echo "could not download javy" && exit 1
fi 
}


( [ $# -eq 0 ] || [ $# -gt 3 ] ) && printUsage && exit 1

case $1 in
  compile)
        [ "`getJavy`" = "could not download javy" ] && echo "could not download javy" && exit 1
        shift 1
        _tsfile=$1

        [ -z "$_tsfile" ] && echo "Error: not specified a ts-file" && printUsage && exit 1
        readlink -e $_tsfile 
        [ $? -ne 0 ] && echo "Error: could not find \"$_tsfile\"" && exit 1

        _jsfile=${_tsfile%.*}.temp.js
        _wasmfile=${2:- ${_tsfile%.*}.wasm}

        npx esbuild ${_tsfile} --bundle --outfile=${_jsfile} --target=es2020 --format=esm
        #convert path to absolute
        tsfile=$(readlink -e $_tsfile)
        [ $? -ne 0 ] && echo "Could not find \"${_tsfile}\"" && exit 1

        temp_js_name=$(mktemp -p ./)
        temp_wasm_name=$(mktemp -p ./)

        cp ${_jsfile} ${temp_js_name}

        #docker run --rm -v ./:/out javy-callback:latest compile -d /out/${temp_js_name} -o /out/${temp_wasm_name}
        ${CACHE_DIR}/${EXECUTABLE} compile -d ${temp_js_name} -o ${temp_wasm_name}

        cp ${temp_wasm_name} ${_wasmfile}
        rm -f ${temp_js_name}
        rm -f ${temp_wasm_name}
    ;;
  emit-provider)
        getJavy
        _wasmfile=$2

        [ -z ${_wasmfile} ] && echo "Error: missing a target wasm file" && printUsage && exit 1; 
        touch ${_wasmfile}
        [ $? -ne 0 ] && echo "Could write to \"$_wasmfile\"" && exit 1

        path=$(readlink -e $_wasmfile)
        path=${path%/*}
        path=${path:-.}
        filename=${_wasmfile##*/}

        #docker run --rm -v ${path}:/out javy-callback:latest emit-provider -o /out/${filename}
        ${CACHE_DIR}/${EXECUTABLE} emit-provider -o ${path}/${filename}

    ;;
  clean)
          rm -rf ${CACHE_DIR}
    ;;
  *)
    echo "Unknown command: $1"
    printUsage
    exit 1
    ;;
esac

