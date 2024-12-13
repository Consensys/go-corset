(defpurefun ((vanishes! :@loob) x) x)
;;
(defcolumns A)
(defpurefun (id x) x)
(defconstraint test () (vanishes! (id A)))
