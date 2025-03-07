;;error:3:33-34:unknown symbol
(defcolumns (X :i16@loob) (Y :i16@loob))
(defconstraint c1 () (reduce + (X Y)))
