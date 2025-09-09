;;error:3:33-40:signed term encountered
(defcolumns (X :u16) (Y :u16))
(deflookup (l1 :unchecked) (X) ((- 1 Y)))
