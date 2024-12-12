(defpurefun ((vanishes! :@loob) x) x)

(defcolumns X ST)
(defconstraint c1 () (vanishes! (* ST (shift X 1))))
