(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns (A :i16) (B :i16))
(defconstraint eq () (vanishes! (- A B)))
(defproperty lem (- A B))
