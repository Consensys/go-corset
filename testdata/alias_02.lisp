(defcolumns X Y)
(defalias
    X' X
    Y' Y)

(defconstraint c1 () (+ X' Y'))
(defconstraint c2 () (+ Y' X'))
(defconstraint c3 () (+ X Y'))
(defconstraint c4 () (+ Y X'))
(defconstraint c5 () (+ X' Y))
(defconstraint c6 () (+ Y' X))
(defconstraint c7 () (+ X Y))
(defconstraint c8 () (+ Y X))
