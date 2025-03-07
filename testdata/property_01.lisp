(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns A B)
(defconstraint eq () (vanishes! (- A B)))
(defproperty lem (- A B))
