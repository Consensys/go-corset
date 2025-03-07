(defpurefun ((vanishes! :𝔽@loob) x) x)
;;
(defcolumns A)
(defpurefun (Id x) x)
(defconstraint test () (vanishes! (Id A)))
