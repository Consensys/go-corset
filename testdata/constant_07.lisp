(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns X ST)
(defconst ONE (^ -2 0))
(defconstraint c1 () (vanishes! (* ST (shift X ONE))))
