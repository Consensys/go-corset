(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns (X :i16))
(defconstraint c1 () (vanishes! (shift X -1)))
