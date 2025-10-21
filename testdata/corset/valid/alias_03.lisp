(defconst
    ONE 1
    TWO 2
)

(defalias
    one ONE
    two TWO)

(defcolumns (X :i16) (Y :i16))
(defconstraint c1 () (== 0 (- X (* two Y))))
(defconstraint c2 () (== 0 (- (* two Y) X)))
(defconstraint c3 () (== 0 (- X Y Y)))
(defconstraint c6 () (== 0 (- (* one X) Y Y)))
