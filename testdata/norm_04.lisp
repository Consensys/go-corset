(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns (ST :i16) (A :i16) (B :i16))
(defconstraint c1 () (vanishes! (* ST (- 1 (~ A)) (- 1 (~ B)))))
