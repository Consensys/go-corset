(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns (X :binary@loob) (Y :binary@bool) A)
(defconstraint c1 () (if X (vanishes! A)))
(defconstraint c2 () (if Y (vanishes! A)))
