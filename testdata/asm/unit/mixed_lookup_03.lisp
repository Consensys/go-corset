(module f)
(defcolumns (X :i32) (Y :i16))
(deflookup (l1 :unchecked) (id.x id.r) (X Y))
