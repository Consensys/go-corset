(defpurefun ((vanishes! :𝔽@loob) x) x)


(defcolumns (X :i16) (Y :i16))
(defalias
    X' X
    Y' Y)

(defconstraint c1 () (vanishes! (- X' Y')))
(defconstraint c2 () (vanishes! (- Y' X')))
(defconstraint c3 () (vanishes! (- X Y')))
(defconstraint c4 () (vanishes! (- Y X')))
(defconstraint c5 () (vanishes! (- X' Y)))
(defconstraint c6 () (vanishes! (- Y' X)))
(defconstraint c7 () (vanishes! (- X Y)))
(defconstraint c8 () (vanishes! (- Y X)))
