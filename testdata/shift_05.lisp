(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns X Y)
(defconstraint c1 () (vanishes! (* (shift X 1) Y)))
