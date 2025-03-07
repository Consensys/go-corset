(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns ST A)
(defconstraint c1 () (vanishes! (* ST (- 1 (~ A)))))
