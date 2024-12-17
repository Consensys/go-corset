(defcolumns (X :@loob) (Y :@loob))
(defconstraint c1 () (reduce + (begin X Y)))
(defconstraint c2 () (reduce + (begin Y X)))
