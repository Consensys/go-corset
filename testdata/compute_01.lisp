(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns (X :i16))
(defcomputed (Y) (id X))
(defconstraint c1 () (vanishes! (- X Y)))
(defconstraint c2 () (vanishes! (- Y X)))
