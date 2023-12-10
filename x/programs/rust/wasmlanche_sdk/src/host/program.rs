//! The `program` module provides functions for calling other programs.
use crate::memory::to_smart_ptr;
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
#[must_use]
pub(crate) fn call(
    caller: &Program,
    target: &Program,
    max_units: i64,
    function_name: &str,
    args: &[u8],
) -> i64 {
    let caller_id = caller.id();
    let caller = to_smart_ptr(&caller_id).unwrap();
    let target_id = target.id();
    let target = to_smart_ptr(&target_id).unwrap();
    let function = to_smart_ptr(function_name.as_bytes()).unwrap();
    let args = to_smart_ptr(args).unwrap();

    unsafe { _call_program(caller, target, max_units, function, args) }
}
