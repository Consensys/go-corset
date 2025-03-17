(defcolumns (A :i16))
(defpurefun ((vanishes! :i16) e) e)
(defconstraint test () (vanishes! A))
