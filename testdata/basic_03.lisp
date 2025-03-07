(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns (X :i16) (Y :i16))
(defconstraint c1 () (vanishes! (- X Y)))
(defconstraint c2 () (vanishes! (- Y X)))
