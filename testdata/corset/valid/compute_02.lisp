(defcolumns (P :i2) (X :i16))
(defcomputed (Y) (filter X P))
(defconstraint c1 (:guard P) (== 0 (- X Y)))
(defconstraint c2 (:guard P) (== 0 (- Y X)))
