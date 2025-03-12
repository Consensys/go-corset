(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns (X :i16) (ST :i16))
(defconstraint c1 () (vanishes! (* ST (shift X 1))))
