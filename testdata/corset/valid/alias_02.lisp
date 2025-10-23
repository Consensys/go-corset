(defcolumns (X :i16) (Y :i16))
(defalias
    X' X
    Y' Y)

(defconstraint c1 () (== 0 (- X' Y')))
(defconstraint c2 () (== 0 (- Y' X')))
(defconstraint c3 () (== 0 (- X Y')))
(defconstraint c4 () (== 0 (- Y X')))
(defconstraint c5 () (== 0 (- X' Y)))
(defconstraint c6 () (== 0 (- Y' X)))
(defconstraint c7 () (== 0 (- X Y)))
(defconstraint c8 () (== 0 (- Y X)))
