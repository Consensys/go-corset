(defpurefun ((vanishes! :@loob) x) x)

(defcolumns X)
(defconstraint c1 () (vanishes! (shift X -1)))
