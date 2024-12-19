;;error:3:33-34:unknown symbol
(defcolumns (X :@loob) (Y :@loob))
(defconstraint c1 () (reduce + (X Y)))
