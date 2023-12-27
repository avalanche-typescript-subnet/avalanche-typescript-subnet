//! The `program` module provides functions for calling other programs.
use crate::errors::StateError;
use crate::memory::to_host_ptr;
use crate::program::Program;

#[link(wasm_import_module = "program")]
extern "C" {
    #[link_name = "call_program"]
    fn _call_program(
        caller_id: i64,
        target_id: i64,
        max_units: i64,
        function: i64,
        args_ptr: i64,
    ) -> i64;
}

/// Calls another program `target` and returns the result.
pub(crate) fn call(
    caller: &Program,
    target: &Program,
    max_units: i64,
    function_name: &str,
    args: &[u8],
) -> Result<i64, StateError> {
    let caller = to_host_ptr(caller.id())?;
    let target = to_host_ptr(target.id())?;
    let function = to_host_ptr(function_name.as_bytes())?;
    let args = to_host_ptr(args)?;

    Ok(unsafe { _call_program(caller, target, max_units, function, args) })
}
