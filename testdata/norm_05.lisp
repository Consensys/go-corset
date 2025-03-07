(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns A)
(defconstraint c1 () (vanishes! (~ A)))
