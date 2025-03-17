(defcolumns (X :i16) (Y :i16))
(defconstraint c1 () (reduce - (begin X Y)))
(defconstraint c2 () (reduce - (begin Y X)))
