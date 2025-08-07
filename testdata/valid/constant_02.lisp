(defpurefun (vanishes! x) (== 0 x))

(defconst ONE 1)
(defcolumns (X :i16))
(defconstraint c1 () (vanishes! (* X (- X ONE))))
