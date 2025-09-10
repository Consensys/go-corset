;;error:3:29-36:signed term encountered
(defcolumns (X :u16) (Y :u16))
(deflookup (l1 :unchecked) ((- 1 X)) (Y))
