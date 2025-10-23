(defcolumns (X :i16) (Y :i16))
(defconstraint c1 () (== 0 (- X Y)))
(defconstraint c2 () (== 0 (- Y X)))
