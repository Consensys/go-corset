;;error:3:22-44:expected loobean constraint (found 𝔽)
(defcolumns X Y)
(defconstraint c1 () (reduce + (begin X Y)))
