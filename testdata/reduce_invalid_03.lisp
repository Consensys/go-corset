;;error:3:22-44:expected loobean constraint (found u16)
(defcolumns (X :i16) (Y :i16))
(defconstraint c1 () (reduce + (begin X Y)))
