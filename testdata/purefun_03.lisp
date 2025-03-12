(defcolumns (A :i16))
(defpurefun ((vanishes! :i16@loob) e) e)
(defconstraint test () (vanishes! A))
