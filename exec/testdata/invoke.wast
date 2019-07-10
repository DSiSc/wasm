(module
 (table 0 anyfunc)
 (memory $0 1)
 (export "memory" (memory $0))
 (export "invoke" (func $invoke))
 (func $invoke (; 0 ;) (param $0 i32) (param $1 i32) (result i32)
  (block $label$0
   (br_if $label$0
    (i32.lt_s
     (get_local $0)
     (i32.const 0)
    )
   )
   (return
    (i32.load
     (get_local $1)
    )
   )
  )
  (i32.const 0)
 )
)
