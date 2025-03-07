(defpurefun ((vanishes! :𝔽@loob) x) x)

(defconst
    (ONE :i32) 1
    (TWO :i64) 2
)

(defcolumns (X :i16) (Y :i16))
(defconstraint c1 () (vanishes! (+ X (* TWO Y))))
(defconstraint c2 () (vanishes! (+ (* TWO Y) X)))
(defconstraint c3 () (vanishes! (+ X Y Y)))
(defconstraint c4 () (vanishes! (+ Y X Y)))
(defconstraint c5 () (vanishes! (+ Y Y X)))
(defconstraint c6 () (vanishes! (+ (* ONE X) Y Y)))
(defconstraint c7 () (vanishes! (+ Y (* ONE X) Y)))
(defconstraint c8 () (vanishes! (+ Y Y (* ONE X))))
