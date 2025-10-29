(defcolumns (X :i16))
(defcomputed ((Y :i16)) (id X))
(defconstraint c1 () (== 0 (- X Y)))
(defconstraint c2 () (== 0 (- Y X)))
