(defpurefun ((vanishes! :@loob) x) x)

(defconst
    ONE 1
    TWO 2
)

(defalias
    one ONE
    two TWO)

(defcolumns X Y)
(defconstraint c1 () (vanishes! (+ X (* two Y))))
(defconstraint c2 () (vanishes! (+ (* two Y) X)))
(defconstraint c3 () (vanishes! (+ X Y Y)))
(defconstraint c4 () (vanishes! (+ Y X Y)))
(defconstraint c5 () (vanishes! (+ Y Y X)))
(defconstraint c6 () (vanishes! (+ (* one X) Y Y)))
(defconstraint c7 () (vanishes! (+ Y (* one X) Y)))
(defconstraint c8 () (vanishes! (+ Y Y (* one X))))
