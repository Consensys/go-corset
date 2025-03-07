(defpurefun ((vanishes! :𝔽@loob) x) x)

(defconst (ONE :extern) 1)
(defcolumns (X :i16))
(defconstraint c1 () (vanishes! (* X (- X ONE))))
