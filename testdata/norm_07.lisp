(defpurefun ((vanishes! :@loob) x) x)

(defcolumns A)
(defconstraint c1 () (vanishes! (- 2 (~ A))))
