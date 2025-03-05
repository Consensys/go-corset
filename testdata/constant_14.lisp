(defpurefun ((vanishes! :@loob) x) x)

(defconst (ONE :extern) 1)
(defcolumns X)
(defconstraint c1 () (vanishes! (* X (- X ONE))))
