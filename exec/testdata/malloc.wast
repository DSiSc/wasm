(module
 (type $FUNCSIG$ii (func (param i32) (result i32)))
 (import "env" "malloc" (func $malloc (param i32) (result i32)))
 (table 0 anyfunc)
 (memory $0 1)
 (export "memory" (memory $0))
 (export "get_str" (func $get_str))
 (export "reverse_str" (func $reverse_str))
 (func $get_str (; 1 ;) (result i32)
  (local $0 i32)
  (i32.store8 offset=4
   (tee_local $0
    (call $malloc
     (i32.const 5)
    )
   )
   (i32.const 111)
  )
  (i32.store align=1
   (get_local $0)
   (i32.const 1819043176)
  )
  (get_local $0)
 )
 (func $reverse_str (; 2 ;) (param $0 i32) (param $1 i32)
  (local $2 i32)
  (local $3 i32)
  (local $4 i32)
  (local $5 i32)
  (set_local $2
   (i32.div_s
    (get_local $1)
    (i32.const 2)
   )
  )
  (block $label$0
   (br_if $label$0
    (i32.lt_s
     (get_local $1)
     (i32.const 2)
    )
   )
   (set_local $1
    (i32.add
     (i32.add
      (get_local $0)
      (get_local $1)
     )
     (i32.const -1)
    )
   )
   (set_local $5
    (i32.const 0)
   )
   (loop $label$1
    (set_local $4
     (i32.load8_u
      (tee_local $3
       (i32.add
        (get_local $0)
        (get_local $5)
       )
      )
     )
    )
    (i32.store8
     (get_local $3)
     (i32.load8_u
      (get_local $1)
     )
    )
    (i32.store8
     (get_local $1)
     (get_local $4)
    )
    (set_local $1
     (i32.add
      (get_local $1)
      (i32.const -1)
     )
    )
    (br_if $label$1
     (i32.lt_s
      (tee_local $5
       (i32.add
        (get_local $5)
        (i32.const 1)
       )
      )
      (get_local $2)
     )
    )
   )
  )
 )
)
