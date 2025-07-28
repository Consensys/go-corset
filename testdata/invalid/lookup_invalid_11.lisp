;;error:3:16-23:signed term encountered
(defcolumns (X :u16) (Y :u16))
(deflookup l1 ((- 1 X)) (Y))
