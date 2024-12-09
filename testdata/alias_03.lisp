(defconst
    ONE 1
    TWO 2
)

(defalias
    one ONE
    two TWO)

(defcolumns X Y)
(defconstraint c1 () (+ X (* two Y)))
(defconstraint c2 () (+ (* two Y) X))
(defconstraint c3 () (+ X Y Y))
(defconstraint c4 () (+ Y X Y))
(defconstraint c5 () (+ Y Y X))
(defconstraint c6 () (+ (* one X) Y Y))
(defconstraint c7 () (+ Y (* one X) Y))
(defconstraint c8 () (+ Y Y (* one X)))
