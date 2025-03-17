(defpurefun (vanishes! x) (== 0 x))

(defconst
    ONE 1
    TWO 2
)

(defalias
    one ONE
    two TWO)

(defcolumns (X :i16) (Y :i16))
(defconstraint c1 () (vanishes! (- X (* two Y))))
(defconstraint c2 () (vanishes! (- (* two Y) X)))
(defconstraint c3 () (vanishes! (- X Y Y)))
(defconstraint c6 () (vanishes! (- (* one X) Y Y)))
