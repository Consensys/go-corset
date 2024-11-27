(defcolumns A)
(defpurefun ((vanishes! :@loob) e) e)
(defconstraint test () (vanishes! A))
