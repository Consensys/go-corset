(defpurefun ((vanishes! :@loob) x) x)

(defcolumns A B)
(defconstraint c1 () (vanishes! (~ (+ A B))))
