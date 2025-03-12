(defcolumns (X :i16@loob) (Y :i16@loob))
(defconstraint c1 () (reduce * (begin X Y)))
(defconstraint c2 () (reduce * (begin Y X)))
