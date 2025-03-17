(defpurefun (vanishes! x) (== 0 x))

(defcolumns (X :i16) (ST :i16))
(defconst ONE (^ -2 0))
(defconstraint c1 () (vanishes! (* ST (shift X ONE))))
