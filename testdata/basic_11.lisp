(defpurefun ((vanishes! :@loob) x) x)

(defcolumns _X _Y)
(defconstraint c1 () (vanishes! (- _X _Y)))
(defconstraint c2 () (vanishes! (- _Y _X)))
