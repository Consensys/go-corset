;;error:3:33-34:unknown symbol
(defcolumns (X :i16) (Y :i16))
(defconstraint c1 () (reduce + (X Y)))
