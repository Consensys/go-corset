(defcolumns (X :i16) (Y :i16))
(defconstraint c1 () (== 0 (reduce * (begin X Y))))
(defconstraint c2 () (== 0 (reduce * (begin Y X))))
