(defpurefun ((vanishes! :@loob) x) x)

(defcolumns A)
(defconstraint c1 () (vanishes! (~ A)))
