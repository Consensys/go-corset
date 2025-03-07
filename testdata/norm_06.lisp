(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns A B)
(defconstraint c1 () (vanishes! (~ (+ A B))))
