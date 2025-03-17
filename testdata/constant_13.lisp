(defpurefun (vanishes! x) (== 0 x))

(defconst
    (ONE :extern) 1
    TWO 2
)

(defcolumns (X :i16) (Y :i64))
(defconstraint c1 () (vanishes! (- X (* TWO Y))))
(defconstraint c2 () (vanishes! (- (* TWO Y) X)))
(defconstraint c3 () (vanishes! (- X Y Y)))
(defconstraint c5 () (vanishes! (- (+ Y Y) X)))
(defconstraint c6 () (vanishes! (- (* ONE X) Y Y)))
(defconstraint c7 () (vanishes! (- (+ Y Y) (* ONE X))))
