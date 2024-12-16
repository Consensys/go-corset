(defcolumns (X :@loob) (Y :@loob))
(defconstraint c1 () (reduce + (X Y)))
