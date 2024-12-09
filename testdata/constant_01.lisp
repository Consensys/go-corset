(defconst
    ONE 1
    TWO 2
)

(defcolumns X Y)
(defconstraint c1 () (+ X (* TWO Y)))
(defconstraint c2 () (+ (* TWO Y) X))
(defconstraint c3 () (+ X Y Y))
(defconstraint c4 () (+ Y X Y))
(defconstraint c5 () (+ Y Y X))
(defconstraint c6 () (+ (* ONE X) Y Y))
(defconstraint c7 () (+ Y (* ONE X) Y))
(defconstraint c8 () (+ Y Y (* ONE X)))
