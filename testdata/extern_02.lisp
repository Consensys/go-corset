(defpurefun ((vanishes! :@loob) x) x)

(defextern ONE 1)
(defcolumns X)
(defconstraint c1 () (vanishes! (* X (- X ONE))))
