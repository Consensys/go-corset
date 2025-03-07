;;error:3:22-44:expected loobean constraint (found 𝔽)
(defcolumns (X :i16) (Y :i16))
(defconstraint c1 () (reduce + (begin X Y)))
