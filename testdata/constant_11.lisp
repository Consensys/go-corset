(defpurefun ((vanishes! :𝔽@loob) x) x)

(module m1)
(defconst
    ONE 1
    TWO 2
)

(defcolumns X Y)
(defconstraint c1 () (vanishes! (+ X (* TWO Y))))
(defconstraint c2 () (vanishes! (+ (* TWO Y) X)))
(defconstraint c3 () (vanishes! (+ X Y Y)))
(defconstraint c4 () (vanishes! (+ Y X Y)))
(defconstraint c5 () (vanishes! (+ Y Y X)))
(defconstraint c6 () (vanishes! (+ (* ONE X) Y Y)))
(defconstraint c7 () (vanishes! (+ Y (* ONE X) Y)))
(defconstraint c8 () (vanishes! (+ Y Y (* ONE X))))
