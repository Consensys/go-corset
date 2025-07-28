;;error:3:20-27:signed term encountered
(defcolumns (X :u16) (Y :u16))
(deflookup l1 (X) ((- 1 Y)))
