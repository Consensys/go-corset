(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns X ST)
(defconstraint c1 () (vanishes! (* ST (shift X 1))))
